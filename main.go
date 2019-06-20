package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/williamsharkey/tui-go"
	"github.com/zenwerk/go-wave"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type post struct {
	username string
	message  string
	time     string
}

var posts = []post{
	{username: "john", message: "hi, what's up?", time: "14:41"},
	{username: "jane", message: "not much", time: "14:43"},
}

var i = 0

func click(rs *RecSettings) {
	rs.Rec = !rs.Rec

	if rs.Rec {
		go recNew(rs)
		newpath := filepath.Join(".", "rec")
		os.MkdirAll(newpath, os.ModePerm)

		audioFileName := filepath.Join(newpath, fmt.Sprintf("%03d.wav", i))

		i = i + 1
		waveFile, err := os.Create(audioFileName)
		if err != nil {
			return
		}
		param := wave.WriterParam{
			Out:           waveFile,
			Channel:       1,
			SampleRate:    44100,
			BitsPerSample: 16,
		}

		waveWriter, err := wave.NewWriter(param)
		if err != nil {
			return
		}
	Loop2:
		for {
			select {
			case <-rs.Print:
				//fmt.Println("s received", p)
			case samps := <-rs.Aud:
				_, err := waveWriter.WriteSample16(samps[:])
				if err != nil {
					waveWriter.Close()
					waveFile.Close()
					return
				}
				//fmt.Print(",")
			case <-rs.Complete:
				waveWriter.Close()
				waveFile.Close()
				//fmt.Println("s complete")
				break Loop2
			default:
			}
		}
	} else {
		rs.Kill <- 1

	}
}

type RecSettings struct {
	Stream   *portaudio.Stream
	Slice    []int16
	Rec      bool
	Kill     chan int
	Complete chan int
	Aud      chan [1024]int16
	Print    chan string
}

func test() {

	rs, err := recInit()
	if err != nil {
		return
	}

	//err=rs.Stream.Start()
	//if err!=nil{
	//	return
	//}
	go click(rs)
	time.Sleep(1 * time.Second)
	go click(rs)
	time.Sleep(4 * time.Second)
	go click(rs)
	time.Sleep(1 * time.Second)
	go click(rs)
	time.Sleep(4 * time.Second)
	go click(rs)
	time.Sleep(1 * time.Second)
	go click(rs)
	time.Sleep(4 * time.Second)
	go click(rs)
	time.Sleep(1 * time.Second)
	go click(rs)
	time.Sleep(4 * time.Second)
	//
	//go recNew(settings)
	//
	//go func(){
	//	time.Sleep(1*time.Second)
	//	settings.Kill<-1
	//}()
	//
	//Loop:
	//for {
	//	select {
	//	case p := <-settings.Print:
	//		fmt.Println("received", p)
	//	case <-settings.Aud:
	//		fmt.Print(".")
	//	case <-settings.Complete:
	//		fmt.Println("complete")
	//		break Loop
	//	default:
	//	}
	//}
	//
	//fmt.Println("hello")

}
func main() {
	//test()
	//return

	//kill := make(chan int, 1)
	//msg := make(chan string, 1)
	//rec := make(chan [1024]int16, 1)
	//p, sl, err := recInit()
	//if err != nil {
	//	return
	//}
	//go recNew(kill, msg, p, sl, rec)
	//
	//i := 0
	//z := 0
	//for {
	//
	//	select {
	//	case <-rec:
	//
	//		z = z + 1
	//		fmt.Print(".")
	//		i = (i + 1) % 80
	//		if i == 0 {
	//			fmt.Println()
	//		}
	//		if z == 200 {
	//			kill <- 333
	//		}
	//
	//	case p := <-msg:
	//		fmt.Println(p)
	//		fmt.Printf("z = %d\n", z)
	//	default:
	//	}
	//}
	//
	//return

	root := tui.NewHBox()

	ui, err := tui.New(root)
	if err != nil {
		s := err.Error()
		fmt.Println(s)
		return
	}

	exitBtn := tui.NewButton("exit")

	exitBtn.OnActivated(func(b *tui.Button) { ui.Quit() })

	playBtn := tui.NewButton("play")

	playBtn.OnActivated(func(b *tui.Button) {})

	recBtn := tui.NewButton("rec")

	rs, err := recInit()

	recBtn.OnActivated(func(b *tui.Button) {
		if rs.Rec {
			recBtn.SetText("rec")
		} else {
			recBtn.SetText("REC")
		}
		go click(rs)

	})

	sidebar := tui.NewVBox(
		playBtn,
		recBtn,
		//tui.NewLabel("CHANNELS"),
		//tui.NewLabel("general"),
		//tui.NewLabel("random"),
		//tui.NewLabel(""),
		//tui.NewLabel("DIRECT MESSAGES"),
		//tui.NewLabel("slackbot"),
		tui.NewSpacer(),
		exitBtn,
	)
	sidebar.SetBorder(true)

	history := tui.NewVBox()

	for _, m := range posts {
		history.Append(tui.NewHBox(
			tui.NewLabel(m.time),
			tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", m.username))),
			tui.NewLabel(m.message),
			tui.NewSpacer(),
		))
	}

	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)

	//
	//btn := tui.NewButton("hey")
	//btn.OnActivated(func(b *tui.Button){
	//	fmt.Println("hey")
	//})

	input := tui.NewEntry()
	//input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetTitle("input")

	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	i2 := tui.NewEntry()

	i2.SetSizePolicy(tui.Expanding, tui.Maximum)
	i2b := tui.NewVBox(i2)
	i2b.SetBorder(true)
	i2b.SetTitle("i2b")

	chat := tui.NewVBox(historyBox, inputBox, i2b)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)

	//root := tui.NewHB
	// ox(chat,sidebar)
	root.Append(sidebar)
	root.Append(chat)
	tui.DefaultFocusChain.Set(playBtn, recBtn, exitBtn, input, i2)

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	input.OnSubmit(func(e *tui.Entry) {
		t := e.Text()
		if t == "quit" || t == "exit" {
			ui.Quit()
			return
		}

		if strings.HasPrefix(e.Text(), "play ") {
			history.Append(tui.NewHBox(
				tui.NewLabel(e.Text()),
				tui.NewSpacer(),
			))
			return
			//input.SetText("")

		}
		history.Append(tui.NewHBox(
			tui.NewLabel(time.Now().Format("15:04")),
			tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", "johnny"))),
			tui.NewLabel(e.Text()),
			tui.NewSpacer(),
		))
		input.SetText("")
	})

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
