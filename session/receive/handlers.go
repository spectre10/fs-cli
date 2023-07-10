package receive

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state != webrtc.ICEConnectionStateClosed {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		}
		if state == webrtc.ICEConnectionStateDisconnected {
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
					Metadata: &lib.Metadata{},
				},
				msgChan: make(chan []byte, 128),
			})
			i := len(s.channels) - 1
			s.channels[i].DC = dc
			s.channels[i].DC.OnOpen(func() {
			})
			s.channels[i].DC.OnMessage(func(msg webrtc.DataChannelMessage) {
				if !s.channels[i].MetadataDone {
					err := json.Unmarshal(msg.Data, &s.channels[i].Metadata)
					if err != nil {
						panic(err)
					}
					s.channels[i].MetadataDone = true
					s.channels[i].File, err = os.OpenFile(s.channels[i].Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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
			// var md lib.Metadata
			// err := json.Unmarshal(msg.Data, &md)
			// if err != nil {
			// 	panic(err)
			// }
			// s.Metadata = &md
			// s.File, err = os.OpenFile(s.Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			// if err != nil {
			// 	panic(err)
			// }
			s.sizeDone = true
			var consent string
			for cap(s.channels) == 0 {
			}
			for !s.channels[0].MetadataDone {
			}
			fmt.Printf("Do you want to receive '%s' ? [Y/N] ", s.channels[0].Name)
			fmt.Scanln(&consent)
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
