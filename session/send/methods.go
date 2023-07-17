package send

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

// Connects clients and creates datachannels.
func (s *Session) Connect(paths []string) error {
	err := s.CreateConnection()
	if err != nil {
		return err
	}
	err = s.CreateControlChannel()
	if err != nil {
		return err
	}

	//here len(paths) is number of files to be sent
	for i := 0; i < len(paths); i++ {
		err = s.CreateTransferChannel(paths[i], i)
		if err != nil {
			return err
		}
	}

	err = s.Createoffer()
	if err != nil {
		return err
	}

	fmt.Println("Paste the remote SDP: ")

	var text string

	//take remote SDP in answer
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

// Creates WebRTC PeerConnection.
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

// Creates Datachannel for file transfer.
func (s *Session) CreateTransferChannel(path string, i int) error {
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
	s.channels[i] = &lib.Document{
		Metadata: &metadata,
		Packet:   make([]byte, 4*4096),
		DCdone:   make(chan struct{}, 1),
		File:     file,
		DCclose:  make(chan struct{}, 1),
	}

	//Ordered property maintains the order of the packets while transferring.
	ordered := true
	//mplt means MaxPacketLifeTime.
	//It is the time in Miliseconds during which if the sender does not receive acknowledgement of the packet, it will retransmit.
	mplt := uint16(5000)
	s.channels[i].DC, err = s.peerConnection.CreateDataChannel(fmt.Sprintf("dc%d", i), &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &mplt,
	})
	if err != nil {
		return err
	}

	//first send the metadata.
	s.channels[i].DC.OnOpen(func() {
		md, err := json.Marshal(s.channels[i].Metadata)
		if err != nil {
			panic(err)
		}
		err = s.channels[i].DC.Send(md)
		if err != nil {
			panic(err)
		}
		close(s.channels[i].DCdone)
		atomic.AddInt32(&s.channelsDone, 1)
	})

	//This indicates that transfer is done on this datachannel.
	s.channels[i].DC.OnMessage(func(msg webrtc.DataChannelMessage) {
		if string(msg.Data) == "completed" {
			s.channels[i].DCclose <- struct{}{}
		}
	})
	return nil
}

// Creates control datachannel for communicating consent and signaling.
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
		// indicates that the operation is complete.
		if signal == "1" {
			s.Close(false)
		}
	})
	return nil
}

// Creates offer and encodes it in base64.
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
