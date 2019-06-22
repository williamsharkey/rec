package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/williamsharkey/tui-go"
	"github.com/zenwerk/go-wave"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Recording struct {
	Name    string
	Samples uint32
}

//var Recs []Recording

func loadRecs() (recs []Recording) {
	path := filepath.Join(".", "wavs")
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}
	for _, file := range files {
		n := file.Name()
		ext := filepath.Ext(n)
		name := strings.TrimSuffix(n, ext)
		if err != nil {
			continue
		}
		recs = append(recs, Recording{name, 0})
	}
	return
}

func click(rs *RecSettings) {
	rs.Rec = !rs.Rec

	if rs.Rec {
		go recNew(rs)

		newpath := filepath.Join(".", "wavs")
		err := os.MkdirAll(newpath, os.ModePerm)

		files, err := ioutil.ReadDir(newpath)
		if err != nil {
			return
		}

		var max int64 = 0
		for _, file := range files {
			n := file.Name()
			ext := filepath.Ext(n)
			name := strings.TrimSuffix(n, ext)
			curr, errP := strconv.ParseInt(name, 10, 64)
			if errP != nil {
				continue
			}
			if curr > max {
				max = curr
			}
		}

		fn := fmt.Sprintf("%03d", max+1)
		audioFileName := filepath.Join(newpath, fn+".wav")

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
				waveWriter.WriteSample16(samps[:])
				if err != nil {
					waveWriter.Close()
					waveFile.Close()
					return
				}

			case <-rs.Complete:
				waveWriter.Close()
				waveFile.Close()
				histAppend(rs.RecList, Recording{fn, 0}, rs.UI)
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
	RecList  *tui.Flexlist
	UI       *tui.UI
}

func main() {
	recs := loadRecs()
	rs, err := recInit()
	if err != nil {
		return
	}

	root := tui.NewHBox()

	ui, err := tui.New(root)
	rs.UI = &ui
	if err != nil {
		s := err.Error()
		fmt.Println(s)
		return
	}

	exitBtn := tui.NewButton("exit")

	exitBtn.OnActivated(func(b *tui.Button) { ui.Quit() })

	playBtn := tui.NewButton("play")

	playBtn.OnActivated(func(b *tui.Button) {})

	recList := tui.NewFlexlist()
	rs.RecList = recList // tui.NewVBox()

	for _, m := range recs {
		histAppend(rs.RecList, m, rs.UI)
	}

	//historyScroll := tui.NewScrollArea(rs.History)
	//historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(recList)
	//recList.OnItemActivated(func(f *tui.Flexlist){historyBox.SetFocused(true)})
	historyBox.SetBorder(true)
	historyBox.SetTitle("wavs")

	recBtn := tui.NewButton("rec")

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
		tui.NewSpacer(),
		exitBtn,
	)
	sidebar.SetBorder(true)

	input := tui.NewEntry()
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetTitle("input")

	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(historyBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)

	root.Append(sidebar)
	root.Append(chat)
	tui.DefaultFocusChain.Set(playBtn, recBtn, exitBtn,

		recList,
		input)

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	input.OnSubmit(func(e *tui.Entry) {
		t := e.Text()
		if t == "quit" || t == "exit" {
			ui.Quit()
			return
		}

		histAppend(rs.RecList, Recording{e.Text(), 0}, rs.UI)
		if strings.HasPrefix(e.Text(), "play ") {
			return
		}

		input.SetText("")
	})

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
func histAppend(box *tui.Flexlist, m Recording, u *tui.UI) {
	box.AddItems(m.Name)
	box.Select(box.Length() - 1)
	(*u).Repaint()
}
