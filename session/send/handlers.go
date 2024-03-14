package send

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spectre10/fs-cli/lib"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/pion/webrtc/v3"
)

// Prints the state change.
func (s *Session) handleState() {
	s.PeerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state != webrtc.ICEConnectionStateClosed {
			fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		}
		if state == webrtc.ICEConnectionStateFailed {
			s.close(false, false)
		}
	})
}

// When control datachannel opens.
func (s *Session) handleopen() func() {
	return func() {
		fmt.Println("Channel opened!")

		//Sends the number of files to be transferred.
		err := s.controlChannel.SendText(fmt.Sprintf("%d", len(s.Channels)))
		if err != nil {
			panic(err)
		}

		fmt.Println("Waiting for receiver to accept the transfer...")
		concentCheck := <-s.Consent
		if !concentCheck {
			fmt.Println("\nReceiver denied to receive.")
			s.close(false, false)
			return
		}

		//wait for all the transfer datachannels to be initiallized.
		//atomic is used to avoid race conditions due to the fact that s.ChannelsDone would be incrementing at the same time.
		for atomic.LoadInt32(&s.channelsDone) != int32(len(s.Channels)) {
		}

		p := mpb.New(
			// mpb.WithWidth(60),
			mpb.WithRefreshRate(100 * time.Millisecond),
		)

		s.GlobalStartTime = lib.Start()

		wg := &sync.WaitGroup{}
		wg.Add(len(s.Channels))
		for i := 0; i < len(s.Channels); i++ {
			//This is because in decor.Any's callback function, i's value changes (idk why!)
			doc := s.Channels[i]

			bar := p.AddBar(int64(s.Channels[i].Size),
				mpb.BarFillerClearOnComplete(), //Clears the bar on completion.
				mpb.PrependDecorators(
					//WCSyncSpaceR synchronizes the margin between multiple bars.
					decor.Name(fmt.Sprintf("Sending '%s': ", s.Channels[i].Name), decor.WCSyncSpaceR),

					//clear byte counter on completion.
					// decor.OnComplete(decor.Counters(decor.SizeB1024(0), "% .2f / % .2f", decor.WCSyncSpaceR), ""),

					//display the sent amount
					//decor.SizeB1024 converts the amount into appropriate units of data (KiB,MiB,Gib)
					decor.OnComplete(decor.Any(func(st decor.Statistics) string {
						stats, _ := s.PeerConnection.GetStats().GetDataChannelStats(doc.DC)
						return fmt.Sprintf("% .2f ", decor.SizeB1024(int64(stats.BytesSent-doc.DC.BufferedAmount())))
					}, decor.WCSyncSpaceR), ""),

					//display speed
					decor.OnComplete(decor.Any(func(st decor.Statistics) string {
						amount := float64(st.Current) / 1048576.0
						period := float64(time.Now().UnixMilli()-doc.StartTime) / 1000.0

						//If the clients are disconnected, do not update the speed.
						if s.PeerConnection.ICEConnectionState() == webrtc.ICEConnectionStateDisconnected {
							return fmt.Sprintf("%.2f MiB/s", 0.0)
						}
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
			proxyReader := bar.ProxyReader(s.Channels[i].File)
			s.Channels[i].StartTime = time.Now().UnixMilli()
			go s.SendFile(proxyReader, i, wg)
		}
		p.Wait()
		wg.Wait()
	}
}

func (s *Session) SendFile(proxyReader io.ReadCloser, i int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer proxyReader.Close()

	eof_chan := make(chan struct{}, 1)
	for {
		select {
		case <-eof_chan:
			<-s.Channels[i].DCclose
			return
		default:
			// Only send packet if the Buffered amount is less than the threshold.
			if s.Channels[i].DC.BufferedAmount() < s.bufferThreshold {
				err := s.sendPacket(proxyReader, s.Channels[i])
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
func (s *Session) sendPacket(proxyReader io.ReadCloser, doc *lib.Document) error {
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
func (s *Session) close(closehandler bool, consent bool) {
	//closehandler indicates if the call came from the listener or it was explicitly called.
	//only handle if the function was explicitly called.
	if !closehandler {
		//get the total size of all the files.
		var fileSize uint64 = 0
		for i := 0; i < len(s.Channels); i++ {
			fileSize += s.Channels[i].Size
		}

		s.stop <- struct{}{}
		for i := 0; i < len(s.Channels); i++ {
			err := s.Channels[i].DC.Close()
			if err != nil {
				panic(err)
			}
		}
		s.controlChannel.Close()
		err := s.PeerConnection.Close()
		if err != nil {
			panic(err)
		}

		if consent {
			t, amount, speed := lib.GetStats(fileSize, s.GlobalStartTime)
			s.TimeTakenSeconds = t
			s.AverageSpeedMiB = speed
			s.TotalAmountTransferred = fmt.Sprintf("% .2f", amount)
			s.StatsDone <- struct{}{}
			fmt.Printf("\nStats:\n")
			fmt.Printf("Time Taken: %.2f seconds\n", t)
			fmt.Printf("Total Amount Transferred: % .2f \n", amount)
			fmt.Printf("Average Speed: %.2f MiB/s\n", speed)
		}

		//wait for the receiver to receive the signal of closing the connection.
		//other wise the receiver hangs and disconnects after no response.
		time.Sleep(1 * time.Second)
		fmt.Println("Connection Closed!")
		close(s.Done)
	}
}

// Handle the closing of control channel.
func (s *Session) handleclose() func() {
	return func() {
		s.close(true, true)
	}
}
