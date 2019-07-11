package main

import (
	"fmt"
	"io"

	"time"

	"gitlab.com/gomidi/midi/mid"
	//driver "github.com/gomidi/rtmididrv"
	// for portmidi
	driver "gitlab.com/gomidi/portmididrv"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// This example expects the first input and output port to be connected
// somehow (are either virtual MIDI through ports or physically connected).
// We write to the out port and listen to the in port.
func playnote(r int, n uint8) {
	drv, err := driver.New()
	must(err)

	// make sure to close all open ports at the end
	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	//if len(os.Args) == 2 && os.Args[1] == "list" {
	//	printInPorts(ins)
	//	printOutPorts(outs)
	//	return
	//}

	if len(ins) == 0 || len(outs) == 0 {
		fmt.Println("no ports")
		return
	}
	in, out := ins[len(ins)-1], outs[len(outs)-1]

	must(in.Open())
	must(out.Open())
	_, pipewr := io.Pipe()
	wr := mid.NewWriter(pipewr)

	// listen for MIDI
	//go mid.NewReader().ReadFrom(in)
	//i:=uint8(0)

	wr.SetChannel(0)
	for q := 0; q < r; q++ { // write MIDI to out that passes it to in on which we listen.

		wr.NoteOn(n, 127)
		//if err != nil {
		//	panic(err)
		//}
		time.Sleep(time.Millisecond * 50)
		wr.NoteOff(n)
		time.Sleep(time.Nanosecond)

		//wr.NoteOn(i+12, 100)
		//time.Sleep(time.Millisecond*10)
		//wr.NoteOff(i+12)
		//time.Sleep(time.Millisecond * 200)
		//i=i+7
	}
}

func printPort(port mid.Port) {
	fmt.Printf("[%v] %s\n", port.Number(), port.String())
}

func printInPorts(ports []mid.In) {
	fmt.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
}

func printOutPorts(ports []mid.Out) {
	fmt.Printf("MIDI OUT Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
}
