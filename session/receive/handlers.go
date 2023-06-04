package receive

import (
	"fmt"
	"strconv"

	"github.com/pion/webrtc/v3"
)

// var Msgchan chan []byte

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		if state == webrtc.ICEConnectionStateDisconnected {
			s.done <- struct{}{}
		}
	})
	s.peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		s.dataChannel = dc
		dc.OnOpen(func() {
			fmt.Printf("New Data Channel Opened! '%s' - '%d'\n", dc.Label(), dc.ID())
		})
		dc.OnClose(func() {
			fmt.Println("Closed!")
			s.close(true)
		})
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			if !s.sizeDone {
				msgstr, err := strconv.Atoi(string(msg.Data))
				if err != nil {
					panic(err)
				}
				s.size = uint64(msgstr)
				s.sizeDone = true
			} else {
				s.msgChan <- msg.Data
			}

		})
	})
}

func (s *Session) close(isOnClose bool) {
	close(s.done)
}
