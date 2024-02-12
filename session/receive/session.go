package receive

import (
	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fs-cli/lib"
)

// Receiver's session struct to manage Datachannels, PeerConnection, Go Signaling Channels etc.
type Session struct {
	peerConnection *webrtc.PeerConnection
	
	controlChannel *webrtc.DataChannel //handling consent and metadata
	channels       []struct {
		*lib.Document
		msgChan chan []byte // This is for sending the packets received via the handler function to the io.writer.
	}
	channelsCnt  int32         //number of channels or the size of the above channels array.
	channelsDone int32         //how many channels are initiallized.
	channelsChan chan struct{} //for Signaling when channelsCnt equals ChannelsDone

	gatherDone <-chan struct{} //for waiting until all the ICECandidates are found.
	done       chan struct{}   //when operation ends.

	ConsentChan     chan struct{} //receiving consent
	
	globalStartTime int64         //start time of the transaction
}

// Constructs new session object and returns it with some default values.
func NewSession() *Session {
	return &Session{
		done:        make(chan struct{}),
		ConsentChan: make(chan struct{}),
		channels: make([]struct {
			*lib.Document
			msgChan chan []byte
		}, 0),
		channelsDone: 0,
		channelsChan: make(chan struct{}, 1),
	}
}
