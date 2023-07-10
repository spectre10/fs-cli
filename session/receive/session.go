package receive

import (
	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

type Session struct {
	peerConnection *webrtc.PeerConnection
	controlChannel *webrtc.DataChannel
	// transferChannel *webrtc.DataChannel
	// channels        []*lib.Document
	channels []struct {
		*lib.Document
		msgChan chan []byte
	}

	gatherDone <-chan struct{}
	// state      *webrtc.ICEConnectionState
	done chan struct{}

	consentChan  chan struct{}
	msgChan      chan []byte
	isChanClosed bool

	sizeDone bool
	// *lib.Document

}

func NewSession() *Session {
	return &Session{
		done:        make(chan struct{}),
		msgChan:     make(chan []byte, 128),
		consentChan: make(chan struct{}),
		// Document: &lib.Document{
		// 	Metadata: &lib.Metadata{},
		// },
		channels: make([]struct {
			*lib.Document
			msgChan chan []byte
		}, 0),
		isChanClosed: false,
		sizeDone:     false,
	}
}
