package receive

import (
	"fmt"
	"io"
	"strings"
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

	bar := p.New(int64(s.Size),
		mpb.BarStyle().Rbound("]"),
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("Receiving '%s': ", s.Name)),
			decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
		),
	)
	proxyWriter := bar.ProxyWriter(s.File)
	go s.fileWrite(proxyWriter, err_chan)
	p.Wait()
	return <-err_chan
}

func (s *Session) fileWrite(proxyWriter io.WriteCloser, err_chan chan error) {
	for {
		select {
		case <-s.done:

			proxyWriter.Close()
			s.File.Close()

			err_chan <- nil
			return

		case msg := <-s.msgChan:
			s.receivedBytes += uint64(len(msg))
			if _, err := proxyWriter.Write(msg); err != nil {
				err_chan <- err
			}

			if s.receivedBytes == s.Size {
				s.controlChannel.SendText("Completed")
			}
		}
	}
}
