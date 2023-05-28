package receive

import (
	"fmt"
	"time"

	"github.com/pion/webrtc/v3"
)

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Println()
		fmt.Printf("ICE Connection State has changed: %s\n", state.String())
		fmt.Println()
		// s.state = &state
		// if state == webrtc.ICEConnectionStateFailed {
		// 	s.done <- struct{}{}
		// }
	})
	s.peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnOpen(func() {
			fmt.Printf("New Data Channel Opened! '%s' - '%d'\n", dc.Label(), dc.ID())
			fmt.Println("Sending New Message Every 2 Seconds!")
			for {
				time.Sleep(2 * time.Second)
				fmt.Println("Sendig new message!")
				dc.SendText("Hello from receiver")
			}
		})
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("New message from data channel '%s' - '%s'\n", dc.Label(), string(msg.Data))
		})
		dc.OnClose(func() {
			fmt.Println("Channel closed!")
			s.done <- struct{}{}
		})
	})
}
