package lib

import "github.com/pion/webrtc/v3"

func Find() ([]string, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(config)
	var address []string
	if err != nil {
		return address, err
	}
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i != nil {
			if i.Typ == webrtc.ICECandidateTypeSrflx {
				address = append(address, i.Address)
			}
		}
	})
	gatherDone := webrtc.GatheringCompletePromise(peerConnection)
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return address, err
	}
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		return address, err
	}
	<-gatherDone
	return address, nil
}
