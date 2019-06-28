package main

import (
	"fmt"
	"github.com/williamsharkey/rec/go-wave"
	"path/filepath"
)

func playNew(fn string, rs *RecSettings) (err error) {
	//playArr := [1024]float64{}

	audioFileName := filepath.Join(".", "wavs", fn+".wav")

	wavReader, err := wave.NewReader(audioFileName)

	if err != nil {
		return
	}

	samps := wavReader.NumSamples

	s := rs.Play
	err = s.PAStream.Start()
	if err != nil {
		s.Complete <- 1

		return
	}

	s.Print <- "started playing"

	for {

		select {

		case x := <-s.Kill:
			s.PAStream.Stop()
			//if err != nil {
			//	return
			//}
			s.Print <- fmt.Sprintf("killed with %d", x)

			s.Complete <- 1
			return

		default:

		}
		f, err := wavReader.ReadSample16()

		if err != nil {
			s.PAStream.Stop()
			//if err != nil {
			//	return
			//}
			s.Print <- "play err " + err.Error()
			s.Complete <- 1

			return err
		}

		//PlayArr := <- s.Buffer
		copy(rs.PlaySlice[:], f[:])
		errWrite := s.PAStream.Write()
		if errWrite != nil {
			s.PAStream.Stop()
			//if err != nil {
			//	return
			//}
			s.Print <- "play err " + errWrite.Error()
			s.Complete <- 1

			return errWrite
		}

		samps = samps - uint32(len(f))

		if samps <= 0 {
			s.Print <- "play complete because samples ran out"
			s.PAStream.Stop()

			s.Complete <- 1

			return err

		}

	}
}
