package receive

import (
	// "os"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

type Session struct {
	// writer         io.Writer
	peerConnection  *webrtc.PeerConnection
	controlChannel  *webrtc.DataChannel
	transferChannel *webrtc.DataChannel

	gatherDone <-chan struct{}
	state      *webrtc.ICEConnectionState
	done       chan struct{}

	consentChan  chan struct{}
	msgChan      chan []byte
	isChanClosed bool

	sizeDone bool
	*lib.Document
	// size     uint64
	// name     string
	// file     *os.File

	receivedBytes uint64
}

func NewSession() *Session {
	return &Session{
		done:        make(chan struct{}),
		msgChan:     make(chan []byte, 63),
		consentChan: make(chan struct{}),
		Document: &lib.Document{
			Metadata: &lib.Metadata{},
		},
		isChanClosed:  false,
		sizeDone:      false,
		receivedBytes: 0,
	}
}
