package main

import "github.com/williamsharkey/rec/portaudio"

func recInit() (set *RecSettings, err error) {
	const bufLen = 1024

	recSlice := make([]int32, bufLen)
	playSlice := make([]int16, bufLen)

	recPAStream, err := portaudio.OpenDefaultStream(1, 0, 44100, bufLen, recSlice)
	if err != nil {
		return
	}

	playPAStream, err := portaudio.OpenDefaultStream(0, 2, 44100, bufLen, playSlice)
	if err != nil {
		return
	}

	return &RecSettings{
		&AudioChan{false, make(chan int), make(chan int), make(chan [bufLen]int32), make(chan string), recPAStream},
		&AudioChan{false, make(chan int), make(chan int), make(chan [bufLen]int32), make(chan string), playPAStream},
		recSlice,
		playSlice,
		nil,
		nil,
		nil,
	}, err
}
