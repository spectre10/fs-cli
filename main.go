package main

import (
	"fmt"
	"os"
	"sync"

	"encoding/base64"
	"encoding/json"

	"github.com/spectre10/fileshare-cli/http"

	"github.com/pion/webrtc/v3"
)

type session struct {
	peerConnection *webrtc.PeerConnection
	dataChannel    *webrtc.DataChannel

	done         chan struct{}
	disconnected chan struct{}

	doneCheckLock sync.Mutex
	doneCheck     bool
}

func main() {
    sess:=newSession()
    sess.connect()
}

func handleError(err error) {
	panic(err)
}

func log(msg string) {
	fmt.Println(msg)
}

func Encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func Decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, obj)
}

func newSession() *session {
	return &session{
		done:         make(chan struct{}),
		disconnected: make(chan struct{}),
		doneCheck:    false,
	}
}

func (s *session) createConnection() error {
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
	s.handleState()
	return nil
}

func (s *session) handleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", state.String())
		if state == webrtc.ICEConnectionStateDisconnected {
			s.disconnected <- struct{}{}
            os.Exit(0)
		}
        if state==webrtc.ICEConnectionStateFailed {
            os.Exit(0)
        }
	})
}

func (s *session) handleopen() func() {
	return func() {
		fmt.Println("Channel opened!")
		fmt.Println("sending message..")
		err := s.dataChannel.SendText("Hello")
		if err != nil {
			panic(err)
		}
	}
}

func (s *session) close(closehandler bool) {
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

func (s *session) handleclose() func() {
	return func() {
		s.close(true)
	}
}

func (s *session) createChannel() error {
	ordered := true
    mplt := uint16(5000)
	channel, err := s.peerConnection.CreateDataChannel("data", &webrtc.DataChannelInit{
		Ordered: &ordered,
        MaxPacketLifeTime: &mplt,
	})
	if err != nil {
		return err
	}
	s.dataChannel = channel
	s.dataChannel.OnOpen(s.handleopen())
	s.dataChannel.OnClose(s.handleclose())
    s.dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
        fmt.Println("message arrived: ",string(msg.Data))
    })
	return nil
}

func (s *session) createoffer() error {
	offer, err := s.peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}

    err=s.peerConnection.SetLocalDescription(offer)
    if err!=nil {
        return err
    }

    encoded,err:=Encode(offer)
    if err!=nil {
        return err
    }
    fmt.Println(encoded)
    return nil
}

func (s *session) connect() error {
	err := s.createConnection()
	if err != nil {
		return err
	}
	err = s.createChannel()
	if err != nil {
		return err
	}

	sdpchan := http.HTTPSDPServer()
	err = s.createoffer()
    if err!=nil {
        return err
    }

	fmt.Println(`Please, provide the SDP via:
curl localhost:8080/sdp --data "$SDP"`)
    answer:=webrtc.SessionDescription{}
    for {
        if err:=Decode(<-sdpchan,&answer);err==nil{
            break
        }
        fmt.Println("Invalid SDP.")
    }

    err=s.peerConnection.SetRemoteDescription(answer)
    if err!=nil {
        return err
    }

    <-s.done

	return nil
}
