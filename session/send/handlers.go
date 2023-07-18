package send

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spectre10/fileshare-cli/lib"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/pion/webrtc/v3"
)

// Prints the state change.
func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state != webrtc.ICEConnectionStateClosed {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		}
	})
}

// When control datachannel opens.
func (s *Session) Handleopen() func() {
	return func() {
		fmt.Println("Channel opened!")

		//Sends the number of files to be transferred.
		err := s.controlChannel.SendText(fmt.Sprintf("%d", len(s.channels)))
		if err != nil {
			panic(err)
		}

		fmt.Println("Waiting for receiver to accept the transfer...")
		concentCheck := <-s.consent
		if !concentCheck {
			fmt.Println("\nReceiver denied to receive.")
			s.Close(false)
			return
		}

		//wait for all the transfer datachannels to be initiallized.
		//atomic is used to avoid race conditions due to the fact that s.channelsDone would be incrementing at the same time.
		for atomic.LoadInt32(&s.channelsDone) != int32(len(s.channels)) {
		}

		p := mpb.New(
			// mpb.WithWidth(60),
			mpb.WithRefreshRate(100 * time.Millisecond),
		)

		wg := &sync.WaitGroup{}
		wg.Add(len(s.channels))
		for i := 0; i < len(s.channels); i++ {
			//This is because in decor.Any's callback function, i's value changes (idk why!)
			doc := s.channels[i]

			bar := p.AddBar(int64(s.channels[i].Size),
				mpb.BarFillerClearOnComplete(), //Clears the bar on completion.
				mpb.PrependDecorators(
					//WCSyncSpaceR synchronizes the margin between multiple bars.
					decor.Name(fmt.Sprintf("Sending '%s': ", s.channels[i].Name), decor.WCSyncSpaceR),

					//clear byte counter on completion.
					// decor.OnComplete(decor.Counters(decor.SizeB1024(0), "% .2f / % .2f", decor.WCSyncSpaceR), ""),

					//display the sent amount
					//decor.SizeB1024 converts the amount into appropriate units of data (KiB,MiB,Gib)
					decor.OnComplete(decor.Any(func(st decor.Statistics) string {
						stats, _ := s.peerConnection.GetStats().GetDataChannelStats(doc.DC)
						return fmt.Sprintf("% .2f ", decor.SizeB1024(int64(stats.BytesSent-doc.DC.BufferedAmount())))
					}, decor.WCSyncSpaceR), ""),

					//display speed
					decor.OnComplete(decor.Any(func(st decor.Statistics) string {
						amount := float64(st.Current) / 1048576.0
						period := float64(time.Now().UnixMilli()-doc.StartTime) / 1000.0
						return fmt.Sprintf("%.2f MiB/s", amount/period)
					}, decor.WCSyncSpaceR), ""),
				),
				mpb.AppendDecorators(
					//replace Percentage with "Done!" on completion.
					decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "Done!"),
				),
			)

			//proxyReader handles the stats(byte counter, percentage) automatically.
			//it is a wrapper of io.Reader.
			proxyReader := bar.ProxyReader(s.channels[i].File)
			s.channels[i].StartTime = time.Now().UnixMilli()
			go s.sendFile(s.channels[i], proxyReader, i, wg)
		}

		p.Wait()
		wg.Wait()
	}
}

func (s *Session) sendFile(doc *lib.Document, proxyReader io.ReadCloser, i int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer proxyReader.Close()

	eof_chan := make(chan struct{}, 1)
	for {
		select {
		case <-eof_chan:
			<-s.channels[i].DCclose
			return
		default:
			// Only send packet if the Buffered amount is less than the threshold.
			if s.channels[i].DC.BufferedAmount() < s.bufferThreshold {
				err := s.SendPacket(proxyReader, s.channels[i])
				if err != nil {
					// if reached End Of File
					if err == io.EOF {
						eof_chan <- struct{}{}
					} else {
						panic(err)
					}
				}
			}
		}
	}
}

// Sends the packet.
func (s *Session) SendPacket(proxyReader io.ReadCloser, doc *lib.Document) error {
	// Read the file to the packet array of size 16KiB.
	n, err := proxyReader.Read(doc.Packet)
	if err != nil {
		return err
	}

	//slice it if the size is less than 16KiB.
	doc.Packet = doc.Packet[:n]

	err = doc.DC.Send(doc.Packet)
	if err != nil {
		return err
	}

	//make the length of packet array 16KiB again.
	doc.Packet = doc.Packet[:cap(doc.Packet)]
	return nil
}

// Closes the channels.
// Ugly.
func (s *Session) Close(closehandler bool) {
	//closehandler indicates if the call came from the listener or it was explicitly called.
	//only handle if the function was explicitly called.
	if !closehandler {
		s.stop <- struct{}{}
		for i := 0; i < len(s.channels); i++ {
			err := s.channels[i].DC.Close()
			if err != nil {
				panic(err)
			}
		}
		s.controlChannel.Close()
		err := s.peerConnection.Close()
		if err != nil {
			panic(err)
		}

		//wait for the receiver to receive the signal of closing the connection.
		//other wise the receiver hangs and disconnects after no response.
		time.Sleep(1 * time.Second)
		fmt.Println("Connection Closed!")
		close(s.done)
	}
}

// Handle the closing of control channel.
func (s *Session) Handleclose() func() {
	return func() {
		s.Close(true)
	}
}
