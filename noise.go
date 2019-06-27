package main

import (
	"github.com/gordonklaus/portaudio"
	"math/rand"
	"time"
)

func noise() {

	h, err := portaudio.DefaultHostApi()
	if err != nil {
		return
	}

	stream, err := portaudio.OpenStream(portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice), func(out []int32) {
		for i := range out {
			out[i] = int32(rand.Uint32())
			//fmt.Println(int32(rand.Uint32()))
		}
	})
	if err != nil {
		return
	}
	defer stream.Close()
	err = stream.Start()
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	err = stream.Stop()
	if err != nil {
		return
	}

}
