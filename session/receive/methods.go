package receive

import (
	"fmt"
	"strings"

	"github.com/pion/webrtc/v3"
	"github.com/pterm/pterm"
	"github.com/spectre10/fileshare-cli/lib"
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

	area, _ := pterm.DefaultArea.Start()
	err_chan := make(chan error)
	go s.fileWrite(area, err_chan)

	return <-err_chan
}

func (s *Session) fileWrite(area *pterm.AreaPrinter, err_chan chan error) {
	for {
		select {
		case <-s.done:
			s.File.Close()
			area.Stop()
			err_chan <- nil
			return
		case msg := <-s.msgChan:
			s.receivedBytes += uint64(len(msg))
			area.Update(pterm.Sprintf("%.2f/%.2f MBs received", float64(s.receivedBytes)/1048576, float64(s.Size)/1048576))
			if _, err := s.File.Write(msg); err != nil {
				err_chan <- err
			}
			if s.receivedBytes == s.Size {
				s.controlChannel.SendText("Completed")
			}
		}
	}
}
