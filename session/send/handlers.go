package send

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/pion/webrtc/v3"
	"github.com/pterm/pterm"
)

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state != webrtc.ICEConnectionStateClosed {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		}
	})
}

func (s *Session) Handleopen() func() {
	return func() {
		fmt.Println("Channel opened!")

		md, err := json.Marshal(s.Metadata)
		if err != nil {
			panic(err)
		}
		s.controlChannel.Send(md)
		fmt.Println("Waiting for receiver to accept the transfer...")
		concentCheck := <-s.consent
		if concentCheck == false {
			fmt.Println("\nReceiver denied to receive.")
			s.Close(false)
			return
		}
		<-s.transferDone
		go s.sendFile()
		// <-s.stop
	}
}

func (s *Session) sendFile() {
	fmt.Println("sending data..")
	area, _ := pterm.DefaultArea.Start()
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(100*time.Millisecond),
	)

	bar := p.New(int64(s.Size),
		mpb.BarStyle().Rbound("]"),
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("Sending '%s': ", s.Name)),
			decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
		),
	)

	proxyReader := bar.ProxyReader(s.File)
	defer proxyReader.Close()
	defer s.File.Close()
	eof_chan := make(chan struct{})
	for {
		select {
		case <-eof_chan:
			p.Wait()
			return
		default:
			if s.transferChannel.BufferedAmount() < s.bufferThreshold {
				err := s.SendPacket(area, proxyReader)
				if err != nil {
					if err == io.EOF {
						// outer:
						// for {
						// 	select {
						// 	case <-s.stop:
						// 		// area.Update(pterm.Sprintf("%.2f/%.2f MBs sent\n", float64(s.Size)/1048576, float64(s.Size)/1048576))
						// 		break outer
						// 	default:
						// 		// area.Update(pterm.Sprintf("%.2f/%.2f MBs sent\n", float64(s.Size-s.transferChannel.BufferedAmount())/1048576, float64(s.Size)/1048576))
						// 	}

						// }
						area.Stop()
						eof_chan <- struct{}{}
					} else {
						panic(err)
					}
				}
			}
		}
	}
}

func (s *Session) SendPacket(area *pterm.AreaPrinter, proxyReader io.ReadCloser) error {
	n, err := proxyReader.Read(s.Packet)
	if err != nil {
		return err
	}

	s.Packet = s.Packet[:n]
	err = s.transferChannel.Send(s.Packet)

	s.Packet = s.Packet[:cap(s.Packet)]
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) Close(closehandler bool) {
	// s.isClosedMut.Lock()
	// if s.isClosed {
	// 	s.isClosedMut.Unlock()
	// 	return
	// }
	if !closehandler {
		s.stop <- struct{}{}
		s.transferChannel.Close()
		s.controlChannel.Close()
		err := s.peerConnection.Close()
		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
		fmt.Println("Connection Closed!")
		close(s.done)
	}
	// fmt.Println("Channel Closed!")
	// time.Sleep(1000 * time.Millisecond)
	// s.stop <- struct{}{}
	// s.isClosed = true
	// s.isClosedMut.Unlock()
}

func (s *Session) Handleclose() func() {
	return func() {
		s.Close(true)
	}
}

func HandleError(err error) {
	panic(err)
}
