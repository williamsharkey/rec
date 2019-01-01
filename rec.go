package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/zenwerk/go-wave"
	"math"
	"os"
	"os/signal"
	"strconv"
	"unsafe"
)

func errCheck(err error) {

	if err != nil {
		panic(err)
	}
}

func effect() {

	r, err := wave.NewReader("test.wav")
	errCheck(err)
	b, err := r.ReadRawSample()
	errCheck(err)
	errCheck(err)
	//samps,err:=r.ReadSampleInt()

	//errCheck(err)
	fmt.Println(len(b))
	fmt.Println("exit")
}

func main() {

	effect()
	return
	cutoff := 5
	if len(os.Args) > 1 {
		fmt.Println(os.Args[1])
		sec, err := strconv.ParseInt(os.Args[1], 10, 32)
		if err == nil {
			cutoff = int(sec)

		}

		if os.Args[1] == "r" {
			effect()
			return
		}
	}

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt, os.Kill)

	audioFileName := "test.wav" //os.Args[1]

	fmt.Println("Recording to "+audioFileName, cutoff, "seconds")

	waveFile, err := os.Create(audioFileName)
	errCheck(err)

	inputChannels := 1
	outputChannels := 0
	sampleRate := 44100
	buf := 4410
	int16Slice := make([]int16, buf)

	portaudio.Initialize()
	//defer portaudio.Terminate()

	apis, err := portaudio.HostApis()

	for i, api := range apis {

		fmt.Println(i, api.Name, api.Type)
		for j, d := range api.Devices {

			fmt.Println(" ", j, d.Name, d.MaxInputChannels, d.MaxOutputChannels)
		}
	}

	stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, float64(sampleRate), len(int16Slice), int16Slice)
	errCheck(err)
	//defer stream.Close()

	// setup Wave file writer

	param := wave.WriterParam{
		Out:           waveFile,
		Channel:       inputChannels,
		SampleRate:    sampleRate,
		BitsPerSample: 16, // if 16, change to WriteSample16()
	}

	waveWriter, err := wave.NewWriter(param)
	errCheck(err)

	defer waveWriter.Close()

	// start reading from microphone
	errCheck(stream.Start())
	fmt.Println("Recording is live now. Say something to your microphone!")
	i := 0

	uInt16Slice := make([]uint16, buf, buf)
	for {

		errCheck(stream.Read())

		i++

		uInt16Slice = *(*[]uint16)(unsafe.Pointer(&int16Slice))
		_, err := waveWriter.WriteSample16(uInt16Slice)

		errCheck(err)
		elapsed := (i * buf) / sampleRate
		fmt.Printf("\r %d", elapsed)
		if cutoff != 0 && elapsed >= cutoff {
			close(waveWriter, stream)
		}
		select {
		case <-sig:
			close(waveWriter, stream)
		default:
		}
	}

	errCheck(stream.Stop())
}

func close(waveWriter *wave.Writer, stream *portaudio.Stream) {
	fmt.Println("\nclosed")

	waveWriter.Close()
	stream.Close()
	portaudio.Terminate()
	os.Exit(0)
}

func conv(a *[]int16, b *[]uint16) {

	for i, q := range *a {

		x := int32(q - math.MinInt16)

		(*b)[i] = uint16(x)

	}

}
