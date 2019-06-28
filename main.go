package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/williamsharkey/tui-go-copy"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	err := portaudio.Initialize()
	if err != nil {
		return
	}
	defer portaudio.Terminate()

	recs := loadRecs()
	rs, err := recInit()
	if err != nil {
		return
	}

	//play("064")
	//clickPlay("064",func(t string){fmt.Print(t)},rs)
	//return
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

	cmdList := tui.NewList()
	cmdList.AddItems("noise", "play", "rec", "nums", "exit")
	cmdList.Select(0)
	cmdList.OnItemActivated(func(l *tui.List) {
		switch l.SelectedItem() {
		case "noise":
			go noise()
		case "play":
			go play(rs.RecList.SelectedItem())
		case "exit":
			exit(ui)
		case "nums":
			go clickNum(rs)
		default:
		}
	})

	cmdBox := tui.NewVBox(cmdList)
	cmdBox.SetBorder(true)

	exitBtn := tui.NewButton("exit")

	exitBtn.OnActivated(func(b *tui.Button) { exit(ui) })

	playBtn := tui.NewButton("play")

	recList := tui.NewList()

	recList.OnItemActivated(func(l *tui.List) {
		go play(l.SelectedItem())
	})

	rs.RecList = recList

	playBtn.OnActivated(func(b *tui.Button) {
		//if rs.Play.Active {
		//	b.SetText("play")
		//} else {
		//	b.SetText("PLAY")
		//}
		go play(rs.RecList.SelectedItem())
		//go clickPlay(rs.RecList.SelectedItem(), func(t string) { b.SetText(t) }, rs)

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

	hbox.Append(cmdBox)
	hbox.Append(sidebar)
	hbox.Append(chat)
	tui.DefaultFocusChain.Set(cmdList, playBtn, recBtn, exitBtn,
		recList,
		input)

	ui.SetKeybinding("Esc", func() { exit(ui) })

	input.OnSubmit(func(e *tui.Entry) {
		t := e.Text()
		if t == "quit" || t == "exit" {
			exit(ui)
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

func exit(ui tui.UI) {
	sig <- os.Kill
	ui.Quit()
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

type RecSettings struct {
	Rec, Play *AudioChan
	RecSlice  []int32
	PlaySlice []int16
	RecList   *tui.List
	UI        *tui.UI
}
type AudioChan struct {
	Active   bool
	Kill     chan int
	Complete chan int
	Buffer   chan [1024]int32
	Print    chan string
	PAStream *portaudio.Stream
}

func histAppend(box *tui.List, u *tui.UI, m ...string) {
	box.AddItems(m...)
	box.Select(box.Length() - 1)
	(*u).Repaint()
}
