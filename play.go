package main

import (
	"github.com/gordonklaus/portaudio"
	"github.com/williamsharkey/go-wave"
	"math"
	"path/filepath"
	"time"
)

func play(fn string) {

	audioFileName := filepath.Join(".", "wavs", fn+".wav")

	wavReader, err := wave.NewReader(audioFileName)
	//fmt.Println(wavReader.FmtChunk)
	if err != nil {
		return
	}
	//
	//for {
	//	framesPerBuffer, err := wavReader.ReadSample()
	//	if err != nil {
	//		break
	//	}
	//	fmt.Println(framesPerBuffer)
	//}
	//if wavReader.NumSamples != wavReader.ReadSampleNum {
	//	fmt.Printf("Actual samples : %d\nRead samples : %d\n", wavReader.NumSamples, wavReader.ReadSampleNum)
	//	fmt.Println(wavReader.NumSamples, wavReader.ReadSampleNum)
	//} else {
	//	fmt.Println("Wave file played without error")
	//
	//}

	h, err := portaudio.DefaultHostApi()
	if err != nil {
		return
	}
	//framesPerBuffer := make([]float64,1)
	//var stream *portaudio.Stream
	//kill:= make(chan int,10)

	b := []int32{}

	for {
		framesPerBuffer, err := wavReader.ReadSample32()
		if err != nil {
			break
		}
		b = append(b, framesPerBuffer[0])
		if wavReader.ReadSampleNum == wavReader.NumSamples {
			break
		}

	}

	q := 0
	//var stream *portaudio.Stream
	stream, err := portaudio.OpenStream(portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice), func(out []int32) {
		//maxsamp:=float64(0)
		//minsamp:=float64(0)
		for i := 0; i < len(out)/2; i = i + 1 {

			//framesPerBuffer, err = wavReader.ReadSample()
			//if err!=nil {
			//
			//
			//	out[i*2] = 0
			//	out[i*2+1]=0
			//} else {
			//const q = float64(3276800.0)
			//if framesPerBuffer[0] > maxsamp{
			//	maxsamp=framesPerBuffer[0]
			//}
			//if framesPerBuffer[0] < minsamp{
			//	minsamp=framesPerBuffer[0]
			//}
			if (i + q) < len(b) {
				out[i*2] = b[(i+q)%len(b)] //float32(framesPerBuffer[0]-1.0)
				out[i*2+1] = out[i*2]

			} else {
				if (i + q) == len(b) {
					//stream.Stop()
					//if err != nil {
					//	return
					//}
				}
				out[i*2] = 0
				out[i*2+1] = 0
			}

			//out[i*2] = b[(i+q)%len(b)] //float32(framesPerBuffer[0]-1.0)
			//out[i*2+1] = out[i*2]             //out[i]
			//fmt.Println(out[i])

			//}

		}
		q = q + len(out)/2
		//fmt.Printf("max samp %f\n",maxsamp)
		//fmt.Printf("min samp %f\n",minsamp)

	})

	if err != nil {
		return
	}
	defer stream.Close()
	err = stream.Start()
	if err != nil {
		return
	}

	//h, err := portaudio.DefaultHostApi()
	//if err != nil {
	//	return
	//}
	//
	//stream, err := portaudio.OpenStream(portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice), func(out []int32) {
	//	wavReader.ReadSample()
	//	for i := range out {
	//		out[i] = int32(rand.Uint32())
	//	}
	//})
	//if err != nil {
	//	return
	//}
	//defer stream.Close()
	//err = stream.Start()
	//if err != nil {
	//	return
	//}

	time.Sleep(time.Millisecond * time.Duration(int(math.Ceil(float64(len(b))/44.1))))
	//fmt.Println("complete")

	err = stream.Stop()
	if err != nil {
		return
	}

}
