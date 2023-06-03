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
	s.isClosedMut.Lock()
	if s.isClosed {
		s.isClosedMut.Unlock()
		return
	}
	if !closehandler {
		s.dataChannel.Close()
	}
	fmt.Println("Channel Closed!")
	time.Sleep(1000 * time.Millisecond)
	s.isClosed = true
	s.isClosedMut.Unlock()

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
