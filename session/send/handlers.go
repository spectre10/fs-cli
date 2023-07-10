package send

import (
	"fmt"
	"io"
	"time"

	"github.com/spectre10/fileshare-cli/lib"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/pion/webrtc/v3"
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

		// md, err := json.Marshal(s.Metadata)
		// if err != nil {
		// 	panic(err)
		// }
		// err = s.controlChannel.Send(md)
		// if err != nil {
		// 	panic(err)
		// }
		err := s.controlChannel.SendText("hello")
		if err != nil {
			panic(err)
		}
		fmt.Println("Waiting for receiver to accept the transfer...")
		concentCheck := <-s.consent
		if !concentCheck {
			fmt.Println("\nReceiver denied to receive.")
			s.Close(false)
			return
		}
		<-s.channels[0].DCdone
		go s.sendFile(s.channels[0])
		// <-s.stop
	}
}

func (s *Session) sendFile(doc *lib.Document) {
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(100*time.Millisecond),
	)

	bar := p.New(int64(doc.Size),
		mpb.BarStyle().Rbound("]"),
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("Sending '%s': ", doc.Name)),
			decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
		),
	)

	proxyReader := bar.ProxyReader(doc.File)
	defer proxyReader.Close()
	eof_chan := make(chan struct{})
	for {
		select {
		case <-eof_chan:
			p.Wait()
			return
		default:
			if s.channels[0].DC.BufferedAmount() < s.bufferThreshold {
				err := s.SendPacket(proxyReader, s.channels[0])
				if err != nil {
					if err == io.EOF {
						eof_chan <- struct{}{}
					} else {
						panic(err)
					}
				}
			}
		}
	}
}

func (s *Session) SendPacket(proxyReader io.ReadCloser, doc *lib.Document) error {
	n, err := proxyReader.Read(doc.Packet)
	if err != nil {
		return err
	}

	doc.Packet = doc.Packet[:n]
	err = doc.DC.Send(doc.Packet)
	if err != nil {
		return err
	}
	doc.Packet = doc.Packet[:cap(doc.Packet)]
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
		// dc.Close()
		for i := 0; i < len(s.channels); i++ {
			_ = s.channels[i].DC.Close()
			// if err!=nil {
			// 	panic(err)
			// }
		}
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
