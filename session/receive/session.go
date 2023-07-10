package receive

import (
	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

type Session struct {
	peerConnection *webrtc.PeerConnection
	controlChannel *webrtc.DataChannel
	channels       []struct {
		*lib.Document
		msgChan chan []byte
	}
	channelsCnt  int32
	channelsDone int32
	channelsChan chan struct{}

	gatherDone <-chan struct{}
	// state      *webrtc.ICEConnectionState
	done chan struct{}

	consentChan  chan struct{}
	isChanClosed bool

	sizeDone bool
	// *lib.Document

}

func NewSession() *Session {
	return &Session{
		done:        make(chan struct{}),
		consentChan: make(chan struct{}),
		// Document: &lib.Document{
		// 	Metadata: &lib.Metadata{},
		// },
		channels: make([]struct {
			*lib.Document
			msgChan chan []byte
		}, 0),
		channelsCnt:  100,
		channelsDone: 0,
		channelsChan: make(chan struct{}, 1),

		isChanClosed: false,
		sizeDone:     false,
	}
}
