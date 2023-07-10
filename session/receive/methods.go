package receive

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func (s *Session) CreateConnection() error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return err
	}
	s.peerConnection = peerConnection
	s.gatherDone = webrtc.GatheringCompletePromise(s.peerConnection)
	s.HandleState()
	return nil
}

func (s *Session) Connect() error {
	err := s.CreateConnection()
	if err != nil {
		return err
	}

	fmt.Println("Paste the SDP:")
	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	offer := webrtc.SessionDescription{}

	err = lib.Decode(input, &offer)
	if err != nil {
		panic(err)
	}
	err = s.peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	answer, err := s.peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	err = s.peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	<-s.gatherDone
	sdp, err := lib.Encode(s.peerConnection.LocalDescription())
	if err != nil {
		panic(err)
	}
	fmt.Println(sdp)

	<-s.consentChan

	err_chan := make(chan error)
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(100*time.Millisecond),
	)
	<-s.channelsChan
	wg := &sync.WaitGroup{}
	wg.Add(int(s.channelsCnt))
	for i := 0; i < int(s.channelsCnt); i++ {
		bar := p.AddBar(int64(s.channels[i].Size),
			// mpb.BarStyle().Rbound("]"),
			mpb.PrependDecorators(
				decor.Name(fmt.Sprintf("Receiving '%s': ", s.channels[i].Name)),
				decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
			),
		)
		proxyWriter := bar.ProxyWriter(s.channels[i].File)
		go s.fileWrite(proxyWriter, err_chan, wg, i)
	}
	p.Wait()
	wg.Wait()
	return nil
}

func (s *Session) fileWrite(proxyWriter io.WriteCloser, err_chan chan error, wg *sync.WaitGroup, i int) {
	var receivedBytes uint64 = 0
	for {
		select {
		case <-s.done:
			wg.Done()
			return
		case msg := <-s.channels[i].msgChan:
			receivedBytes += uint64(len(msg))
			if _, err := proxyWriter.Write(msg); err != nil {
				panic(err)
			}

			if receivedBytes == s.channels[i].Size {
				err := proxyWriter.Close()
				if err != nil {
					panic(err)
				}
				err = s.controlChannel.SendText("1")
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
