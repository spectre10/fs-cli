package receive

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"

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
	s.peerConnection = peerConnection
	s.gatherDone = webrtc.GatheringCompletePromise(s.peerConnection)
	s.HandleState()
	return nil
}

// Connects the clients and starts the process of writing to the file.
func (s *Session) Connect() error {
	err := s.CreateConnection()
	if err != nil {
		return err
	}

	fmt.Println("Paste the SDP:")
	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	offer := webrtc.SessionDescription{}

	// Decode from base64 to SDP
	err = lib.Decode(input, &offer)
	if err != nil {
		panic(err)
	}
	err = s.peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	answer, err := s.peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	err = s.peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	<-s.gatherDone
	
	//Encode the SDP to base64
	sdp, err := lib.Encode(s.peerConnection.LocalDescription())
	if err != nil {
		panic(err)
	}
	fmt.Println(sdp)

	<-s.consentChan

	// Initialize new mpb instance.
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(100*time.Millisecond), //updates the stats every 100ms.
	)

	//wait for all the channels to be initialized.
	<-s.channelsChan

	wg := &sync.WaitGroup{}
	wg.Add(int(s.channelsCnt))
	for i := 0; i < int(s.channelsCnt); i++ {
		bar := p.AddBar(int64(s.channels[i].Size),
			mpb.BarFillerClearOnComplete(), // Make the progress bar disappear on completion.
			mpb.PrependDecorators(
				decor.Name(fmt.Sprintf("Receiving '%s': ", s.channels[i].Name), decor.WCSyncSpaceR),
				//Make the size counter disapper on completion.
				decor.OnComplete(decor.Counters(decor.SizeB1024(0), "% .2f / % .2f", decor.WCSyncSpaceR), ""),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "Done!"), //Replace percentage with "Done!" on completion.
			),
		)
		//mpb's proxyWriter to automatically handle the progress bar and stats
		proxyWriter := bar.ProxyWriter(s.channels[i].File)
		go s.fileWrite(proxyWriter, wg, i)
	}
	
	//wait for all the bars to complete.
	p.Wait()
	//wait for all the fileWrite functions to complete.
	wg.Wait()
	//signal to the sender to close the connection.
	err = s.controlChannel.SendText("1")
	if err != nil {
		panic(err)
	}
	<-s.done
	return nil
}

func (s *Session) fileWrite(proxyWriter io.WriteCloser, wg *sync.WaitGroup, i int) {
	var receivedBytes uint64 = 0
	signalChan := make(chan struct{}, 1)
	for {
		select {
		case <-signalChan:
			//signal the completion of a particular file.
			err := s.channels[i].DC.SendText("completed")
			if err != nil {
				panic(err)
			}
			wg.Done()
			return
		case msg := <-s.channels[i].msgChan:
			receivedBytes += uint64(len(msg))
			//write packet
			if _, err := proxyWriter.Write(msg); err != nil {
				panic(err)
			}

			//If all the packets are received, close the writer and go to the first case of select.
			if receivedBytes == s.channels[i].Size {
				err := proxyWriter.Close()
				if err != nil {
					panic(err)
				}
				signalChan <- struct{}{}
			}
		}
	}
}
