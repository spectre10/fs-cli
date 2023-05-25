package session

import (
	"fmt"
	"github.com/pion/webrtc/v3"
	"strings"
	"sync"
	// "github.com/spectre10/fileshare-cli/http"
	"github.com/spectre10/fileshare-cli/lib"
)

type Session struct {
	peerConnection *webrtc.PeerConnection
	dataChannel    *webrtc.DataChannel

	done         chan struct{}
	disconnected chan struct{}

	doneCheckLock sync.Mutex
	doneCheck     bool
}

func NewSession() *Session {
	return &Session{
		done:         make(chan struct{}),
		disconnected: make(chan struct{}),
		doneCheck:    false,
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
	s.HandleState()
	return nil
}

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", state.String())
		if state == webrtc.ICEConnectionStateDisconnected {
			s.disconnected <- struct{}{}
		} else if state == webrtc.ICEConnectionStateFailed {
			s.done <- struct{}{}
		}
	})
}

func (s *Session) Handleopen() func() {
	return func() {
		fmt.Println("Channel opened!")
		fmt.Println("sending message..")
		err := s.dataChannel.SendText("Hello")
		if err != nil {
			panic(err)
		}
	}
}

func (s *Session) Close(closehandler bool) {
	if closehandler == false {
		s.dataChannel.Close()
	}

	s.doneCheckLock.Lock()
	if s.doneCheck == true {
		s.doneCheckLock.Unlock()
		return
	}
	s.doneCheck = true
	s.doneCheckLock.Unlock()

	close(s.done)
}

func (s *Session) Handleclose() func() {
	return func() {
		s.Close(true)
	}
}

func (s *Session) CreateChannel() error {
	// ordered := true
	// mplt := uint16(5000)
	channel, err := s.peerConnection.CreateDataChannel("data", &webrtc.DataChannelInit{
		// Ordered:           &ordered,
		// MaxPacketLifeTime: &mplt,
	})
	if err != nil {
		return err
	}
	s.dataChannel = channel
	s.dataChannel.OnOpen(s.Handleopen())
	s.dataChannel.OnClose(s.Handleclose())
	s.dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Println("message arrived: ", string(msg.Data))
	})
	return nil
}

func (s *Session) Createoffer() error {
	offer, err := s.peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}

	err = s.peerConnection.SetLocalDescription(offer)
	if err != nil {
		return err
	}

	encoded, err := lib.Encode(offer)
	if err != nil {
		return err
	}
	fmt.Println(encoded)
	return nil
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

	// sdpchan := http.HTTPSDPServer()

	err = s.Createoffer()
	if err != nil {
		return err
	}

	fmt.Println("Paste the remote SDP: ")

	var text string
	fmt.Scanln(&text)
	text = strings.TrimSpace(text)
	sdp := text

	answer := webrtc.SessionDescription{}
	for {
		if err := lib.Decode(sdp, &answer); err == nil {
			break
		}
		fmt.Println("Invalid SDP.")
	}

	err = s.peerConnection.SetRemoteDescription(answer)
	if err != nil {
		return err
	}

	<-s.done

	return nil
}

func HandleError(err error) {
	panic(err)
}

func log(msg string) {
	fmt.Println(msg)
}
