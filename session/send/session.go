package send

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

type Session struct {
	peerConnection  *webrtc.PeerConnection
	control         *webrtc.DataChannel
	dataChannel     *webrtc.DataChannel
	bufferThreshold uint64

	done       chan struct{}
	gatherDone <-chan struct{}
	stop       chan struct{}

	data   []byte
	reader io.Reader

	size uint64
	name string

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
	return &Session{
		done:            make(chan struct{}),
		data:            make([]byte, 4*4096),
		bufferThreshold: 1024 * 1024,
		stop:            make(chan struct{}),
		reader:          file,
		size:            uint64(f.Size()),
		name:            f.Name(),
		consent:         make(chan bool),
		consentDone:     false,
	}
}

func (s *Session) Connect() error {
	err := s.CreateConnection()
	if err != nil {
		return err
	}
	err = s.CreateChannel()
	if err != nil {
		return err
	}

	err = s.Createoffer()
	if err != nil {
		return err
	}

	fmt.Println("Paste the remote SDP: ")

	var text string

	answer := webrtc.SessionDescription{}
	for {
		text, err = lib.ReadSDP()
		if err != nil {
			return err
		}
		sdp := text
		if err := lib.Decode(sdp, &answer); err == nil {
			break
		}
		fmt.Println("Invalid SDP. Enter again.")
	}

	err = s.peerConnection.SetRemoteDescription(answer)
	if err != nil {
		return err
	}

	<-s.done

	return nil
}

func log(msg string) {
	fmt.Println(msg)
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
	s.HandleState()
	return nil
}

func (s *Session) CreateChannel() error {
	ordered := true
	mplt := uint16(5000)
	channel, err := s.peerConnection.CreateDataChannel("control", &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &mplt,
	})
	if err != nil {
		return err
	}
	s.dataChannel = channel
	s.dataChannel.OnOpen(s.Handleopen())
	s.dataChannel.OnClose(s.Handleclose())
	s.dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		if s.consentDone == false {
			if string(msg.Data) == "n" {
				s.consent <- false
				s.consentDone = true
				return
			}
			s.consent <- true
			s.consentDone = true
			return
		}
		signal := string(msg.Data)
		if signal == "Completed" {
			s.Close(false)
			s.stop <- struct{}{}
		}
	})
	return nil
}

func (s *Session) Createoffer() error {
	offer, err := s.peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}
	s.gatherDone = webrtc.GatheringCompletePromise(s.peerConnection)
	err = s.peerConnection.SetLocalDescription(offer)
	<-s.gatherDone
	offer2 := s.peerConnection.LocalDescription()
	if err != nil {
		return err
	}

	encoded, err := lib.Encode(offer2)
	if err != nil {
		return err
	}
	fmt.Println(encoded)
	return nil
}
