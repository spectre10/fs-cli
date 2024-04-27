package receive

import (
	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fs-cli/lib"
)

// Receiver's session struct to manage Datachannels, PeerConnection, Go Signaling Channels etc.
type Session struct {
	PeerConnection *webrtc.PeerConnection

	controlChannel *webrtc.DataChannel //handling consent and metadata
	Channels       []struct {
		*lib.Document
		msgChan chan []byte // This is for sending the packets received via the handler function to the io.writer.
	}
	channelsCnt  int32         //number of channels or the size of the above channels array.
	channelsDone int32         //how many channels are initiallized.
	channelsChan chan struct{} //for Signaling when channelsCnt equals ChannelsDone

	gatherDone <-chan struct{} //for waiting until all the ICECandidates are found.
	done       chan struct{}   //when operation ends.

	consentChan   chan struct{} //receiving consent
	ConsentInput  chan byte
	MetadataReady chan struct{} //when metadata is received.

	// stats at the end
	GlobalStartTime int64 //start time of the transaction
	TimeTakenSeconds       float64
	AverageSpeedMiB        float64
	TotalAmountTransferred string
	StatsDone              chan struct{}
}

// Constructs new session object and returns it with some default values.
func NewSession() *Session {
	return &Session{
		done:          make(chan struct{}),
		consentChan:   make(chan struct{}),
		ConsentInput:  make(chan byte, 1),
		MetadataReady: make(chan struct{}, 1),
		Channels: make([]struct {
			*lib.Document
			msgChan chan []byte
		}, 0),
		channelsDone: 0,
		channelsChan: make(chan struct{}, 1),
		StatsDone:    make(chan struct{}, 1),
	}
}
