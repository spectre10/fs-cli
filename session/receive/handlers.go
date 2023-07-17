package receive

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

// Handle all the listeners.
func (s *Session) HandleState() {
	//print the state change
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state != webrtc.ICEConnectionStateClosed {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		}
		if state == webrtc.ICEConnectionStateDisconnected {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
			s.done <- struct{}{}
		}
	})

	//On new DataChannel being created by sender.
	s.peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		if dc.Label() == "control" {
			s.controlChannel = dc
			//add listeners to control channel.
			s.assign(dc)
		} else {
			//append new file and construct Document struct.
			s.channels = append(s.channels, struct {
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

			i := len(s.channels) - 1
			s.channels[i].DC = dc
			s.channels[i].DC.OnClose(func() {
				// fmt.Println("Channel", dc.Label(), "Closed")
			})
			s.channels[i].DC.OnMessage(func(msg webrtc.DataChannelMessage) {
				//first receive metadata of files(name,size).
				if !s.channels[i].MetadataDone {
					err := json.Unmarshal(msg.Data, &s.channels[i].Metadata)
					if err != nil {
						panic(err)
					}
					s.channels[i].MetadataDone = true
					s.channels[i].File, err = os.OpenFile(s.channels[i].Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

					//increment the channelsDone counter
					//atomic write is used avoid race conditions due to multiple channels being intiallized at once.
					atomic.AddInt32(&s.channelsDone, 1)
					if err != nil {
						panic(err)
					}
				} else {
					s.channels[i].msgChan <- msg.Data
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

		//take consent from receiver.
		var consent string
		for i := 0; i < len(s.channels); i++ {
			fmt.Printf(" %s ", s.channels[i].Name)
		}
		fmt.Printf("\nDo you want to receive the above files? [Y/N] ")
		fmt.Scanln(&consent)
		fmt.Println()

		//send appropriate consent to sender.
		if consent == "n" || consent == "N" {
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
