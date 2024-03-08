package receive

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fs-cli/lib"
)

// Handle all the listeners.
func (s *Session) HandleState() {
	//print the state change
	s.PeerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateFailed {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
			s.done <- struct{}{}
			return
		}

		if state != webrtc.ICEConnectionStateClosed {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		}
	})

	//On new DataChannel being created by sender.
	s.PeerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		if dc.Label() == "control" {
			s.controlChannel = dc
			//add listeners to control channel.
			s.assign(dc)
		} else {
			//append new file and construct Document struct.
			s.Channels = append(s.Channels, struct {
				*lib.Document
				msgChan chan []byte
			}{
				Document: &lib.Document{
					Metadata: &lib.Metadata{},
				},
				// initiallize msgChan to have 128 * sizeOfPacket
				// 128 * (16384 Bytes) = 2MiB of buffer-like storage if the speed of incoming packets is more than the write speed.
				msgChan: make(chan []byte, 128),
			})

			i := len(s.Channels) - 1
			s.Channels[i].DC = dc
			s.Channels[i].DC.OnClose(func() {
				// fmt.Println("Channel", dc.Label(), "Closed")
			})
			s.Channels[i].DC.OnMessage(func(msg webrtc.DataChannelMessage) {
				//first receive metadata of files(name,size).
				if !s.Channels[i].MetadataDone {
					err := json.Unmarshal(msg.Data, &s.Channels[i].Metadata)
					if err != nil {
						panic(err)
					}
					s.Channels[i].MetadataDone = true
					s.Channels[i].File, err = os.OpenFile(s.Channels[i].Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

					//increment the channelsDone counter
					//atomic write is used avoid race conditions due to multiple channels being initiallized at once.
					atomic.AddInt32(&s.channelsDone, 1)
					if err != nil {
						panic(err)
					}
				} else {
					s.Channels[i].msgChan <- msg.Data
				}
			})
		}
	})
}

// Handle listeners of control channel.
func (s *Session) assign(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		// fmt.Printf("New Data Channel Opened! '%s' - '%d'\n", dc.Label(), dc.ID())
	})
	dc.OnClose(func() {
		fmt.Println("Connection Closed!")
		s.close(true)
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		cnt, err := strconv.Atoi(string(msg.Data))
		//due to atomic not having int type.
		cnt32 := int32(cnt)
		s.channelsCnt = atomic.LoadInt32(&cnt32)
		if err != nil {
			panic(err)
		}

		//wait for all the data-channels to be initiallized.
		for atomic.LoadInt32(&s.channelsCnt) != atomic.LoadInt32(&s.channelsDone) {
		}
		s.channelsChan <- struct{}{}
		s.MetadataReady <- struct{}{}

		//take consent from receiver.
		consent:= <-s.ConsentInput

		//send appropriate consent to sender.
		if consent == 'n' || consent == 'N' {
			err := s.controlChannel.SendText("n")
			if err != nil {
				panic(err)
			}
		} else {
			err := s.controlChannel.SendText("Y")
			if err != nil {
				panic(err)
			}
			s.consentChan <- struct{}{}
		}
	})
}

// Closes all the go channels.(effectively closing the operation)
func (s *Session) close(isOnClose bool) {
	close(s.consentChan)
	close(s.done)
}
