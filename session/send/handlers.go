package send

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
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
		err := s.controlChannel.SendText(fmt.Sprintf("%d", len(s.channels)))
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
		for atomic.LoadInt32(&s.channelsDone) != int32(len(s.channels)) {
		}
		p := mpb.New(
			mpb.WithWidth(60),
			mpb.WithRefreshRate(100*time.Millisecond),
		)
		wg := &sync.WaitGroup{}
		wg.Add(len(s.channels))
		for i := 0; i < len(s.channels); i++ {
			bar := p.AddBar(int64(s.channels[i].Size), mpb.BarFillerClearOnComplete(),
				// mpb.BarStyle().Rbound("]"),
				mpb.PrependDecorators(
					decor.Name(fmt.Sprintf("Sending '%s': ", s.channels[i].Name), decor.WCSyncSpaceR),
					decor.OnComplete(decor.Counters(decor.SizeB1024(0), "% .2f / % .2f", decor.WCSyncSpaceR), ""),
				),
				mpb.AppendDecorators(
					decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
					// decor.OnComplete(decor.Counters(decor.SizeB1024(0), "% .2f / % .2f", decor.WCSyncSpaceR), "fuck"),
				),
			)
			proxyReader := bar.ProxyReader(s.channels[i].File)
			go s.sendFile(s.channels[i], proxyReader, i, wg)
		}
		wg.Wait()
		p.Wait()
		// <-s.stop
	}
}

func (s *Session) sendFile(doc *lib.Document, proxyReader io.ReadCloser, i int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer proxyReader.Close()
	eof_chan := make(chan struct{})
	for {
		select {
		case <-eof_chan:
			<-s.channels[i].DCclose
			return
		default:
			if s.channels[i].DC.BufferedAmount() < s.bufferThreshold {
				err := s.SendPacket(proxyReader, s.channels[i])
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
