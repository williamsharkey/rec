package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/williamsharkey/tui-go-copy"
	"github.com/zenwerk/go-wave"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {

	recs := loadRecs()
	rs, err := recInit()
	if err != nil {
		return
	}

	//clickPlay("044",func(t string){fmt.Print(t)},rs)
	//clickRec(rs)

	hbox := tui.NewHBox()
	root := tui.NewPadder(1, 0, hbox)

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

	recList := tui.NewList()
	rs.RecList = recList

	playBtn.OnActivated(func(b *tui.Button) {
		if rs.Play.Active {
			b.SetText("play")
		} else {
			b.SetText("PLAY")
		}
		go clickPlay(rs.RecList.SelectedItem(), func(t string) { b.SetText(t) }, rs)

	})

	histAppend(rs.RecList, rs.UI, recs...)

	historyBox := tui.NewVBox(recList)

	historyBox.SetBorder(true)
	historyBox.SetTitle("wavs")

	recBtn := tui.NewButton("rec")

	recBtn.OnActivated(func(b *tui.Button) {
		if rs.Rec.Active {
			b.SetText("rec")
		} else {
			b.SetText("REC")
		}
		go clickRec(rs)

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

	hbox.Append(sidebar)
	hbox.Append(chat)
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

		histAppend(rs.RecList, rs.UI, e.Text())
		if strings.HasPrefix(e.Text(), "play ") {
			return
		}

		input.SetText("")
	})

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}

func loadRecs() (recs []string) {
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
		recs = append(recs, name)
	}
	return
}

func clickRec(rs *RecSettings) {
	s := rs.Rec
	s.Active = !s.Active

	if rs.Rec.Active {
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
			case <-rs.Rec.Print:
				//fmt.Println("s received", p)
			case samps := <-rs.Rec.Buffer:
				waveWriter.WriteSample16(samps[:])
				if err != nil {
					waveWriter.Close()
					waveFile.Close()
					return
				}

			case <-rs.Rec.Complete:
				waveWriter.Close()
				waveFile.Close()
				histAppend(rs.RecList, rs.UI, fn)
				break Loop2
			default:
			}
		}
	} else {
		rs.Rec.Kill <- 1
	}
}

func clickPlay(fn string, setBtn func(string), rs *RecSettings) {

	if fn == "" {
		return
	}
	s := rs.Play
	s.Active = !s.Active

	if s.Active {
		go playNew(fn, rs)

	LoopPlay:
		for {
			select {
			case p := <-s.Print:
				fmt.Println("received", p)
			//case samps := <-rs.Rec.Buffer:
			//	waveWriter.WriteSample16(samps[:])
			//	if err != nil {
			//		waveWriter.Close()
			//		waveFile.Close()
			//		return
			//	}

			case t := <-s.Complete:
				fmt.Println("play complete", t)
				setBtn("play")
				s.Active = false
				//waveWriter.Close()
				//waveFile.Close()
				//histAppend(rs.RecList, rs.UI, fn)
				break LoopPlay
			default:
			}
		}
	} else {
		s.Kill <- 1
	}
}

type RecSettings struct {
	Rec, Play *AudioChan
	RecSlice  []int16
	PlaySlice []int16
	RecList   *tui.List
	UI        *tui.UI
}
type AudioChan struct {
	Active   bool
	Kill     chan int
	Complete chan int
	Buffer   chan [1024]int16
	Print    chan string
	PAStream *portaudio.Stream
}

func histAppend(box *tui.List, u *tui.UI, m ...string) {
	box.AddItems(m...)
	box.Select(box.Length() - 1)
	(*u).Repaint()
}
