package receive

import (
	"fmt"
	"io"
	"strings"

	"github.com/pion/webrtc/v3"
	"github.com/pterm/pterm"
	"github.com/spectre10/fileshare-cli/lib"
)

type Session struct {
	writer         io.Writer
	peerConnection *webrtc.PeerConnection
	dataChannel    *webrtc.DataChannel

	gatherDone <-chan struct{}
	state      *webrtc.ICEConnectionState
	done       chan struct{}

	msgChan      chan []byte
	isChanClosed bool

	size     uint64
	sizeDone bool

	receivedBytes uint64
}

func NewSession(file io.Writer) *Session {
	return &Session{
		done:          make(chan struct{}),
		writer:        file,
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
			area.Stop()
			return nil
		case msg := <-s.msgChan:
			s.receivedBytes += uint64(len(msg))
			area.Update(pterm.Sprintf("%f/%f MBs received", float32(s.receivedBytes)/1000000, float32(s.size)/1000000))
			if _, err := s.writer.Write(msg); err != nil {
				fmt.Println(err)
			}
			if s.receivedBytes == s.size {
				s.dataChannel.SendText("Completed")
			}
		}
	}
	// <-s.done
	// return nil
}
