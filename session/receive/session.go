package receive

import (
	"fmt"
	// "io"
	"os"
	"strings"

	"github.com/pion/webrtc/v3"
	"github.com/pterm/pterm"
	"github.com/spectre10/fileshare-cli/lib"
)

type Session struct {
	// writer         io.Writer
	peerConnection *webrtc.PeerConnection
	dataChannel    *webrtc.DataChannel

	gatherDone <-chan struct{}
	state      *webrtc.ICEConnectionState
	done       chan struct{}

	msgChan      chan []byte
	isChanClosed bool

	size     uint64
	sizeDone bool
	name     string
	file     *os.File

	receivedBytes uint64
}

func NewSession() *Session {
	return &Session{
		done:          make(chan struct{}),
		msgChan:       make(chan []byte),
		isChanClosed:  false,
		size:          0,
		sizeDone:      false,
		receivedBytes: 0,
	}
}

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
	area, _ := pterm.DefaultArea.Start()
	for {
		select {
		case <-s.done:
			s.file.Close()
			area.Stop()
			return nil
		case msg := <-s.msgChan:
			s.receivedBytes += uint64(len(msg))
			area.Update(pterm.Sprintf("%f/%f MBs received", float64(s.receivedBytes)/1048576, float64(s.size)/1048576))
			if _, err := s.file.Write(msg); err != nil {
				fmt.Println(err)
			}
			if s.receivedBytes == s.size {
				s.dataChannel.SendText("Completed")
			}
		}
	}
}
