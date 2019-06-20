package main

import (
	"fmt"
	//"github.com/gordonklaus/portaudio"
	//"fmt"
	//"github.com/gordonklaus/portaudio"
	"github.com/gordonklaus/portaudio"
	"github.com/zenwerk/go-wave"
	"os"
	//"os"
	//"os/signal"
)

func Record(cutoff int, name string, f func(s string), turnOff func(), errf func(s string)) (killRec chan int) {

	killRec = make(chan int, 1)

	recImpl(cutoff, name, killRec, f, turnOff, errf)
	return
}

func recImpl(cutoff int, name string, killRec chan int, f func(s string), turnOff func(), errf func(s string)) (err error) {

	//sig := make(chan os.Signal, 1)

	//	signal.Notify(sig, os.Interrupt, os.Kill)

	audioFileName := name + ".wav" //os.Args[1]

	//fmt.Println("Recording to "+audioFileName, cutoff, "seconds")

	waveFile, err := os.Create(audioFileName)
	if err != nil {

		errf(err.Error())
		return
	}

	inputChannels := 1
	outputChannels := 0
	sampleRate := 44100
	buf := 4410
	int16Slice := make([]int16, buf)

	portaudio.Initialize()
	//defer portaudio.Terminate()

	//apis, err := portaudio.HostApis()

	//for i, api := range apis {
	//
	//	fmt.Println(i, api.Name, api.Type)
	//	for j, d := range api.Devices {
	//
	//		fmt.Println(" ", j, d.Name, d.MaxInputChannels, d.MaxOutputChannels)
	//	}
	//}

	stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, float64(sampleRate), len(int16Slice), int16Slice)
	if err != nil {
		errf(err.Error())
		return
	}
	//defer stream.Close()

	// setup Wave file writer

	param := wave.WriterParam{
		Out:           waveFile,
		Channel:       inputChannels,
		SampleRate:    sampleRate,
		BitsPerSample: 16, // if 16, change to WriteSample16()
	}

	waveWriter, err := wave.NewWriter(param)
	if err != nil {
		errf(err.Error())
		return
	}
	defer waveWriter.Close()

	// start reading from microphone

	err = stream.Start()
	if err != nil {
		errf(err.Error())
		return
	}
	//fmt.Println("Recording is live now. Say something to your microphone!")
	i := 0

	//uInt16Slice := make([]uint16, buf, buf)
	for {

		err = stream.Read()
		if err != nil {
			errf(err.Error())
			return
		}

		i++

		//uInt16Slice = *(*[]uint16)(unsafe.Pointer(&int16Slice))
		_, err := waveWriter.WriteSample16(int16Slice)

		if err != nil {
			errf(err.Error())
			return err
		}
		elapsed := (i * buf) / sampleRate
		f(fmt.Sprintf("%d", elapsed))
		//fmt.Printf("\r %d", elapsed)
		if cutoff != 0 && elapsed >= cutoff {
			closeX(waveWriter, stream)
			turnOff()
			return err
		}
		select {
		//case <-sig:
		//	close(sig)
		//	close(killRec)
		//	closeX(waveWriter, stream)
		//	turnOff()
		//	return err
		case <-killRec:
			//close(sig)
			close(killRec)
			fmt.Println("kill")
			closeX(waveWriter, stream)
			turnOff()
			return err

		default:
		}
	}

	err = stream.Stop()
	if err != nil {
		errf(err.Error())
		return err
	}
	return
}

func closeX(waveWriter *wave.Writer, stream *portaudio.Stream) {
	//fmt.Println("\nclosed")

	waveWriter.Close()
	stream.Close()
	portaudio.Terminate()
	//os.Exit(0)
}
