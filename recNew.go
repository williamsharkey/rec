package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
)

func recInit() (set *RecSettings, err error) {
	err = portaudio.Initialize()
	if err != nil {
		return
	}
	const buf = 1024
	sl := make([]int16, buf)

	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, buf, sl)
	if err != nil {
		return
	}

	//killRec := make(chan int)
	//complete := make(chan int)
	//recChan := make(chan [1024]int16)
	//print := make(chan string)
	//

	return &RecSettings{
		stream,
		sl,
		false,
		make(chan int),
		make(chan int),
		make(chan [buf]int16),
		make(chan string),
		nil,
		nil,
	}, err
}

func recNew(rs *RecSettings) (err error) {
	recA := [1024]int16{}
	err = rs.Stream.Start()
	if err != nil {
		return
	}
	rs.Print <- "started"
	for {

		select {

		case x := <-rs.Kill:
			err = rs.Stream.Stop()
			if err != nil {
				return
			}
			rs.Print <- fmt.Sprintf("killed with %d", x)

			rs.Complete <- 1
			return

		default:
		}

		err = rs.Stream.Read()
		if err != nil {
			err = rs.Stream.Stop()
			if err != nil {
				return
			}
			rs.Print <- "rec err " + err.Error()
			rs.Complete <- 1

			return
		}
		copy(recA[:], rs.Slice[:])
		rs.Aud <- recA

	}
}
