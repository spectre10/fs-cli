package receive

import (
	"fmt"
	"io"
	"strings"

	"github.com/pion/webrtc/v3"
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
		done:     make(chan struct{}),
		writer:   file,
		msgChan:  make(chan []byte),
		size:     0,
		sizeDone: false,
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

	for {
		select {
		case <-s.done:
			return nil
		case msg := <-s.msgChan:
			s.receivedBytes += uint64(len(msg))
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
