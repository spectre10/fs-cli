package send

import (
	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

// To manage the datachannels and PeerConnection.
type Session struct {
	peerConnection *webrtc.PeerConnection

	controlChannel *webrtc.DataChannel
	controlDone    chan struct{}

	//Maximum amount the buffer can store for each datachannel.
	bufferThreshold uint64

	done       chan struct{}
	gatherDone <-chan struct{}
	stop       chan struct{}

	channels     []*lib.Document
	channelsCnt  int32
	channelsDone int32

	consent     chan bool
	consentDone bool
}

// Returns new Session object with some default values.
func NewSession(numberOfFiles int) *Session {
	return &Session{
		done:            make(chan struct{}),
		bufferThreshold: 512 * 1024, //512KiB
		controlDone:     make(chan struct{}, 1),
		stop:            make(chan struct{}, 1),
		channels:        make([]*lib.Document, numberOfFiles),
		channelsCnt:     0,
		channelsDone:    0,
		consent:         make(chan bool),
		consentDone:     false,
	}
}
