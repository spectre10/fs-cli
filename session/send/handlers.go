package send

import (
    "github.com/pion/webrtc/v3"
    "fmt"
)

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
        fmt.Println()
		fmt.Printf("ICE Connection State has changed: %s\n", state.String())
        fmt.Println()
		// if state == webrtc.ICEConnectionStateDisconnected {
			// s.disconnected <- struct{}{}
		// } 
  //       else if state == webrtc.ICEConnectionStateFailed {
		// 	s.done <- struct{}{}
		// }
	})
}

func (s *Session) Handleopen() func() {
	return func() {
		fmt.Println("Channel opened!")
		fmt.Println("sending message..")
		err := s.dataChannel.SendText("Hello")
		if err != nil {
			panic(err)
		}
	}
}

func (s *Session) Close(closehandler bool) {
	if closehandler == false {
		s.dataChannel.Close()
	}

	s.doneCheckLock.Lock()
	if s.doneCheck == true {
		s.doneCheckLock.Unlock()
		return
	}
	s.doneCheck = true
	s.doneCheckLock.Unlock()

	close(s.done)
}

func (s *Session) Handleclose() func() {
	return func() {
		s.Close(true)
	}
}

func HandleError(err error) {
	panic(err)
}
