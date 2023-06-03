package receive

import (
	"fmt"
	"github.com/pion/webrtc/v3"
)

var Msgchan chan []byte

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		if state == webrtc.ICEConnectionStateFailed {
			s.done <- struct{}{}
		}
	})
	s.peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnOpen(func() {
			fmt.Printf("New Data Channel Opened! '%s' - '%d'\n", dc.Label(), dc.ID())
		})
		dc.OnClose(func() {
			fmt.Println("Closed!")
			s.close(true)
		})
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			s.msgChan <- msg.Data
		})
	})
}

func (s *Session) close(isOnClose bool) {
	close(s.done)
}
