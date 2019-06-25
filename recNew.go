package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/zenwerk/go-wave"
	"path/filepath"
)

func recInit() (set *RecSettings, err error) {
	err = portaudio.Initialize()
	if err != nil {
		return
	}
	const buf = 1024
	recSlice := make([]int16, buf)
	playSlice := make([]int16, buf)

	//recPAStream, err := portaudio.OpenDefaultStream(1, 0, 44100, buf, recSlice)
	//if err != nil {
	//	return
	//}

	playPAStream, err := portaudio.OpenDefaultStream(0, 2, 44100, buf, playSlice)
	if err != nil {
		return
	}

	return &RecSettings{
		&AudioChan{false, make(chan int), make(chan int), make(chan [buf]int16), make(chan string), nil},
		&AudioChan{false, make(chan int), make(chan int), make(chan [buf]int16), make(chan string), playPAStream},
		recSlice,
		playSlice,
		nil,
		nil,
	}, err
}

func recNew(rs *RecSettings) (err error) {
	recArr := [1024]int16{}
	s := rs.Rec
	err = s.PAStream.Start()
	if err != nil {
		return
	}
	rs.Rec.Print <- "started recording"
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

		err = s.PAStream.Read()
		if err != nil {
			s.PAStream.Stop()
			//if err != nil {
			//	return
			//}
			s.Print <- "rec err " + err.Error()
			s.Complete <- 1

			return
		}
		copy(recArr[:], rs.RecSlice[:])
		s.Buffer <- recArr

	}
}

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
		f, err := wavReader.ReadSampleInt16()

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
		if err == nil {
			samps = samps - uint32(len(f))
		}
		if samps <= 0 {
			s.Print <- "play complete because samples ran out"
			s.PAStream.Stop()

			s.Complete <- 1

			return err

		}

	}
}
