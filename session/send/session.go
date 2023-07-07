package send

import (
	"os"
	"sync"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

type Session struct {
	peerConnection *webrtc.PeerConnection

	transferChannel *webrtc.DataChannel
	transferDone    chan struct{}

	controlChannel *webrtc.DataChannel
	controlDone    chan struct{}

	bufferThreshold uint64

	done       chan struct{}
	gatherDone <-chan struct{}
	stop       chan struct{}

	*lib.Document

	isClosedMut sync.Mutex
	isClosed    bool

	consent     chan bool
	consentDone bool
}

func NewSession(path string) *Session {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	f, err := file.Stat()
	if err != nil {
		panic(err)
	}
	metadata := lib.Metadata{
		Name: f.Name(),
		Size: uint64(f.Size()),
	}
	return &Session{
		done:            make(chan struct{}),
		bufferThreshold: 512 * 1024,
		controlDone:     make(chan struct{}, 1),
		transferDone:    make(chan struct{}, 1),
		stop:            make(chan struct{}, 1),
		Document: &lib.Document{
			Metadata: &metadata,
			Packet:   make([]byte, 4*4096),
			File:     file,
		},
		consent:     make(chan bool),
		consentDone: false,
	}
}
