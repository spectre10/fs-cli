package receive

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fs-cli/lib"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// Creates new WebRTC peerConnection.
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
	s.PeerConnection = peerConnection
	s.gatherDone = webrtc.GatheringCompletePromise(s.PeerConnection)
	s.HandleState()
	return nil
}

// Connects the clients and starts the process of writing to the file.
func (s *Session) Connect(offer webrtc.SessionDescription) error {
	<-s.consentChan

	//wait for all the channels to be initialized.
	<-s.channelsChan

	go s.transfer()
	<-s.done
	return nil
}

func (s *Session) transfer() {

	// Initialize new mpb instance.
	p := mpb.New(
		// mpb.WithWidth(60),
		mpb.WithRefreshRate(100 * time.Millisecond), //updates the stats every 100ms.
	)
	s.globalStartTime = lib.Start()
	wg := &sync.WaitGroup{}
	wg.Add(int(s.channelsCnt))
	for i := 0; i < int(s.channelsCnt); i++ {
		// because i's value changes in decor.Any's callback function.
		doc := s.Channels[i]
		bar := p.AddBar(int64(s.Channels[i].Size),
			mpb.BarFillerClearOnComplete(), // Make the progress bar disappear on completion.
			mpb.PrependDecorators(
				decor.Name(fmt.Sprintf("Receiving '%s': ", s.Channels[i].Name), decor.WCSyncSpaceR),

				//Make the size counter disappear on completion.
				// decor.OnComplete(decor.Counters(decor.SizeB1024(0), "% .2f / % .2f", decor.WCSyncSpaceR), ""),

				//display the received amount
				//decor.SizeB1024 converts the amount into appropriate units of data (KiB,MiB,Gib)
				decor.OnComplete(decor.Any(func(st decor.Statistics) string {
					stats, _ := s.PeerConnection.GetStats().GetDataChannelStats(doc.DC)
					return fmt.Sprintf("% .2f ", decor.SizeB1024(int64(stats.BytesReceived)))
				}, decor.WCSyncSpaceR), ""),

				//display speed
				decor.OnComplete(decor.Any(func(st decor.Statistics) string {
					amount := float64(st.Current) / 1048576.0
					period := float64(time.Now().UnixMilli()-doc.StartTime) / 1000.0

					// If the clients are disconnected, do not update speed.
					if s.PeerConnection.ICEConnectionState() == webrtc.ICEConnectionStateDisconnected {
						return fmt.Sprintf("%.2f MiB/s", 0.0)
					}
					return fmt.Sprintf("%.2f MiB/s", amount/period)
				}, decor.WCSyncSpaceR), ""),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "Done!"), //Replace percentage with "Done!" on completion.
			),
		)
		//mpb's proxyWriter to automatically handle the progress bar and stats
		proxyWriter := bar.ProxyWriter(s.Channels[i].File)
		s.Channels[i].StartTime = time.Now().UnixMilli()
		go s.fileWrite(proxyWriter, wg, i)
	}

	//wait for all the bars to complete.
	p.Wait()
	//wait for all the fileWrite functions to complete.
	wg.Wait()

	//get total size of all the files.
	var fileSize uint64 = 0
	for i := 0; i < int(s.channelsCnt); i++ {
		fileSize += s.Channels[i].Size
	}

	lib.FinalStat(fileSize, s.globalStartTime)

	//signal to the sender to close the connection.
	err := s.controlChannel.SendText("1")
	if err != nil {
		panic(err)
	}
	<-s.done
}

func (s *Session) fileWrite(proxyWriter io.WriteCloser, wg *sync.WaitGroup, i int) {
	var receivedBytes uint64 = 0
	signalChan := make(chan struct{}, 1)
	for {
		select {
		case <-signalChan:
			//signal the completion of a particular file.
			err := s.Channels[i].DC.SendText("completed")
			if err != nil {
				panic(err)
			}
			wg.Done()
			return
		case msg := <-s.Channels[i].msgChan:
			receivedBytes += uint64(len(msg))
			//write packet
			if _, err := proxyWriter.Write(msg); err != nil {
				panic(err)
			}

			//If all the packets are received, close the writer and go to the first case of select.
			if receivedBytes == s.Channels[i].Size {
				err := proxyWriter.Close()
				if err != nil {
					panic(err)
				}
				signalChan <- struct{}{}
			}
		}
	}
}

func (s *Session) GenSDP(offer webrtc.SessionDescription) (string, error) {
	var sdp string
	err := s.PeerConnection.SetRemoteDescription(offer)
	if err != nil {
		return sdp, err
	}

	answer, err := s.PeerConnection.CreateAnswer(nil)
	if err != nil {
		return sdp, err
	}

	err = s.PeerConnection.SetLocalDescription(answer)
	if err != nil {
		return sdp, err
	}
	<-s.gatherDone

	//Encode the SDP to base64
	sdp, err = lib.Encode(s.PeerConnection.LocalDescription())
	return sdp, err
}

func (s *Session) PrintSDP(offer webrtc.SessionDescription) error {
	sdp, err := s.GenSDP(offer)
	if err != nil {
		return err
	}
	fmt.Println(sdp)
	return nil

}
