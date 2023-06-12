package send

import (
	"fmt"
	"github.com/pion/webrtc/v3"
	"github.com/pterm/pterm"
	"io"
	"time"
)

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
	})
}

func (s *Session) Handleopen() func() {
	return func() {
		fmt.Println("Channel opened!")
		fmt.Println("sending data..")
		s.dataChannel.SendText(fmt.Sprint(s.size))
		area, _ := pterm.DefaultArea.Start()
		for {
			select {
			case <-s.stop:
				return
			default:
				err := s.SendPacket(area)
				if err != nil {
					if err == io.EOF {
						// s.stop <- struct{}{}
						area.Stop()
					} else {
						panic(err)
					}
				}
			}
		}
	}
}

func (s *Session) SendPacket(area *pterm.AreaPrinter) error {
	n, err := s.reader.Read(s.data)
	if err != nil {
		// s.Close(false)
		return err
	}
	s.data = s.data[:n]
	err = s.dataChannel.Send(s.data)
	s.data = s.data[:cap(s.data)]
	if err != nil {
		return err
	}
	stats, _ := s.peerConnection.GetStats().GetDataChannelStats(s.dataChannel)
	area.Update(pterm.Sprintf("%f/%f MBs sent", float64(stats.BytesSent)/1048576, float64(s.size)/1048576))

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
