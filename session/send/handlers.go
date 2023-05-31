package send

import (
	"fmt"
	"io"
	"time"

	"github.com/pion/webrtc/v3"
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
		for {
			select {
			case <-s.stop:
				return
			default:
				err := s.SendPacket()
				if err != nil {
					if err == io.EOF {
						s.stop <- struct{}{}
						// return
					} else {
						panic(err)
					}
				}
			}
		}
	}
}

func (s *Session) SendPacket() error {
	n, err := s.reader.Read(s.data)
	if err != nil {
		s.Close(false)
		return err
	}
	s.data = s.data[:n]
	err = s.dataChannel.Send(s.data)
	s.data = s.data[:cap(s.data)]
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) Close(closehandler bool) {
	if closehandler == false {
		s.dataChannel.Close()
		fmt.Println("data channel closed")
		time.Sleep(5 * time.Second)
	}
	// s.doneCheckLock.Lock()
	// if s.doneCheck == true {
	// s.doneCheckLock.Unlock()
	// return
	// }
	// s.doneCheck = true
	// s.doneCheckLock.Unlock()

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
