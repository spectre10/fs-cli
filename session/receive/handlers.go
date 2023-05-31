package receive

import (
	"fmt"
	// "math/rand"
	// "time"

	"github.com/pion/webrtc/v3"
)

var Msgchan chan []byte

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Println()
		fmt.Printf("ICE Connection State has changed: %s\n", state.String())
		fmt.Println()
		// if state == webrtc.ICEConnectionStateDisconnected {
		// 	s.done <- struct{}{}
		// }
		// s.state = &state
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
			// fmt.Printf("New message from data channel '%s' - '%s'\n", dc.Label(), string(msg.Data))

			// _, err := s.writer.Write(msg.Data)
			s.msgChan <- msg.Data
			// if err != nil {
			// 	fmt.Println(err)
			// }
		})
	})
}

func (s *Session) close(isOnClose bool) {
	// if isOnClose == false {
	// 	s.dataChannel.Close()
	// }
	// s.chanClosedMut.Lock()
	// if s.chanClosed {
	// 	s.chanClosedMut.Unlock()
	// 	return
	// }
	// s.chanClosed = true
	// s.chanClosedMut.Unlock()
	// s.done <- struct{}{}
	close(s.done)
}
