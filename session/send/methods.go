package send

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fs-cli/lib"
)

func (s *Session) SetupConnection(paths []string) error {
	err := s.createConnection()
	if err != nil {
		return err
	}
	err = s.createControlChannel()
	if err != nil {
		return err
	}

	//here len(paths) is number of files to be sent
	for i := 0; i < len(paths); i++ {
		err = s.createTransferChannel(paths[i], i)
		if err != nil {
			return err
		}
	}
	return nil
}

// Connects clients.
func (s *Session) Connect(answer webrtc.SessionDescription) error {
	err := s.PeerConnection.SetRemoteDescription(answer)
	if err != nil {
		return err
	}

	<-s.done

	return nil
}

// Creates WebRTC PeerConnection.
func (s *Session) createConnection() error {
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
	s.PeerConnection = peerConnection
	s.handleState()
	return nil
}

// Creates Datachannel for file transfer.
func (s *Session) createTransferChannel(path string, i int) error {
	var err error
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

	//create new metadata struct.
	metadata := lib.Metadata{
		Name: f.Name(),
		Size: uint64(f.Size()),
	}
	//create new document struct.
	s.Channels[i] = &lib.Document{
		Metadata: &metadata,
		Packet:   make([]byte, 4*4096),
		DCdone:   make(chan struct{}, 1),
		File:     file,
		DCclose:  make(chan struct{}, 1),
	}

	//Ordered property maintains the order of the packets while transferring.
	ordered := true
	//mplt means MaxPacketLifeTime.
	//It is the time in Milliseconds during which if the sender does not receive acknowledgement of the packet, it will retransmit.
	mplt := uint16(5000)
	s.Channels[i].DC, err = s.PeerConnection.CreateDataChannel(fmt.Sprintf("dc%d", i), &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &mplt,
	})
	if err != nil {
		return err
	}

	//first send the metadata.
	s.Channels[i].DC.OnOpen(func() {
		md, err := json.Marshal(s.Channels[i].Metadata)
		if err != nil {
			panic(err)
		}
		err = s.Channels[i].DC.Send(md)
		if err != nil {
			panic(err)
		}
		close(s.Channels[i].DCdone)
		atomic.AddInt32(&s.channelsDone, 1)
	})

	//This indicates that transfer is done on this datachannel.
	s.Channels[i].DC.OnMessage(func(msg webrtc.DataChannelMessage) {
		if string(msg.Data) == "completed" {
			s.Channels[i].DCclose <- struct{}{}
		}
	})
	return nil
}

// Creates control datachannel for communicating consent and signaling.
func (s *Session) createControlChannel() error {
	ordered := true
	mplt := uint16(5000)
	channel, err := s.PeerConnection.CreateDataChannel("control", &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &mplt,
	})
	if err != nil {
		return err
	}
	s.controlChannel = channel
	s.controlChannel.OnOpen(s.handleopen())
	s.controlChannel.OnClose(s.handleclose())
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
		// indicates that the operation is complete.
		if signal == "1" {
			s.close(false)
		}
	})
	return nil
}

func (s *Session) GenOffer() (string, error) {
	offer, err := s.PeerConnection.CreateOffer(nil)
	if err != nil {
		return "", err
	}
	s.gatherDone = webrtc.GatheringCompletePromise(s.PeerConnection)
	err = s.PeerConnection.SetLocalDescription(offer)
	<-s.gatherDone
	offer2 := s.PeerConnection.LocalDescription()

	if err != nil {
		return "", err
	}

	encoded, err := lib.Encode(offer2)
	return encoded, err
}

// Creates offer and encodes it in base64.
func (s *Session) PrintOffer() error {
	offer, err := s.GenOffer()
	if err != nil {
		return err
	}
	fmt.Println(offer)
	return nil
}
