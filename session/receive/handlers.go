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

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state != webrtc.ICEConnectionStateClosed {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		}
		if state == webrtc.ICEConnectionStateDisconnected {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
			s.done <- struct{}{}
		}
		//       else if state == webrtc.ICEConnectionStateConnected {
		// 	// s.close(true)
		// }
	})
	s.peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		if dc.Label() == "control" {
			s.controlChannel = dc
			s.assign(dc)
		} else {
			s.channels = append(s.channels, struct {
				*lib.Document
				msgChan chan []byte
			}{
				Document: &lib.Document{
					Metadata:         &lib.Metadata{},
					MetadataDoneChan: make(chan struct{}, 1),
				},
				msgChan: make(chan []byte, 128),
			})
			i := len(s.channels) - 1
			s.channels[i].DC = dc
			s.channels[i].DC.OnClose(func() {
				// fmt.Println("Channel", dc.Label(), "Closed")
			})
			s.channels[i].DC.OnMessage(func(msg webrtc.DataChannelMessage) {
				if !s.channels[i].MetadataDone {
					err := json.Unmarshal(msg.Data, &s.channels[i].Metadata)
					if err != nil {
						panic(err)
					}
					s.channels[i].MetadataDone = true
					s.channels[i].File, err = os.OpenFile(s.channels[i].Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
					s.channels[i].MetadataDoneChan <- struct{}{}
					atomic.AddInt32(&s.channelsDone, 1)
					// if atomic.LoadInt32(&s.channelsCnt) == atomic.LoadInt32(&s.channelsDone) {
					// 	s.channelsChan <- struct{}{}
					// }
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

func (s *Session) assign(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		// fmt.Printf("New Data Channel Opened! '%s' - '%d'\n", dc.Label(), dc.ID())
	})
	dc.OnClose(func() {
		fmt.Println("Connection Closed!")
		s.close(true)
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		if !s.sizeDone {
			cnt, err := strconv.Atoi(string(msg.Data))
			cnt32 := int32(cnt)
			s.channelsCnt = atomic.LoadInt32(&cnt32)
			if err != nil {
				panic(err)
			}
			s.sizeDone = true
			var consent string
			// <-s.channelsChan
			for atomic.LoadInt32(&s.channelsCnt) != atomic.LoadInt32(&s.channelsDone) {
			}
			s.channelsChan <- struct{}{}
			for i := 0; i < len(s.channels); i++ {
				fmt.Printf(" %s ", s.channels[i].Name)
			}
			fmt.Printf("\nDo you want to receive the above files? [Y/N] ")
			fmt.Scanln(&consent)
			fmt.Println()
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
		}
	})
}

func (s *Session) close(isOnClose bool) {
	close(s.consentChan)
	close(s.done)
}
