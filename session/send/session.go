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
	peerConnection *webrtc.PeerConnection
	dataChannel    *webrtc.DataChannel

	done       chan struct{}
	gatherDone <-chan struct{}
	stop       chan struct{}

	data   []byte
	reader io.Reader

	path string
	size uint64

	isClosedMut sync.Mutex
	isClosed    bool
}

func NewSession(path string) *Session {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	f, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	sz := f.Size()
	return &Session{
		done:      make(chan struct{}),
		data:      make([]byte, 4096),
		stop:      make(chan struct{}),
		reader:    file,
		path:      path,
		size:      uint64(sz),
		// gatherDone: make(chan struct{}),
		// doneCheck: false,
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

	// sdp := http.HTTPSDPServer()

	err = s.Createoffer()
	if err != nil {
		return err
	}

	fmt.Println("Paste the remote SDP: ")

	var text string
	// fmt.Scanln(&text)
	// text = strings.TrimSpace(text)
	// sdp := text

	answer := webrtc.SessionDescription{}
	for {
		text, err = lib.MustReadStdin()
		if err != nil {
			return err
		}
		// fmt.Scanln(&text)
		// text = strings.TrimSpace(text)
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
	channel, err := s.peerConnection.CreateDataChannel("data", &webrtc.DataChannelInit{
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
