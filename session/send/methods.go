package send

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

func (s *Session) Connect(path string) error {
	err := s.CreateConnection()
	if err != nil {
		return err
	}
	err = s.CreateControlChannel()
	if err != nil {
		return err
	}
	err = s.CreateTransferChannel(path)
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

func (s *Session) CreateTransferChannel(path string) error {
	ordered := true
	mplt := uint16(5000)
	var err error
	// s.channels[0] = &lib.Document{
	// 	Metadata: &lib.Metadata{},
	// 	Packet:   make([]byte, 4*4096),
	// }
	// s.channels[0].File, err = os.Open(path)
	if err != nil {
		panic(err)
	}
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
	s.channels[0] = &lib.Document{
		Metadata: &metadata,
		Packet:   make([]byte, 4*4096),
		DCdone:   make(chan struct{}, 1),
		File:     file,
	}
	// s.channels[0].DCdone = make(chan struct{}, 1)
	// s.channels[0].Metadata = &metadata
	// s.channels[0].Packet = make([]byte, 4*4096)
	s.channels[0].DC, err = s.peerConnection.CreateDataChannel("dc0", &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &mplt,
	})

	if err != nil {
		return err
	}
	s.channels[0].DC.OnOpen(func() {
		md, err := json.Marshal(s.channels[0].Metadata)
		if err != nil {
			panic(err)
		}
		err = s.channels[0].DC.Send(md)
		if err != nil {
			panic(err)
		}
		close(s.channels[0].DCdone)
	})
	return nil
	// s.transferChannel.OnOpen()
}

func (s *Session) CreateControlChannel() error {
	ordered := true
	mplt := uint16(5000)
	channel, err := s.peerConnection.CreateDataChannel("control", &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &mplt,
	})
	if err != nil {
		return err
	}
	s.controlChannel = channel
	s.controlChannel.OnOpen(s.Handleopen())
	s.controlChannel.OnClose(s.Handleclose())
	s.controlChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		if !s.consentDone {
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
