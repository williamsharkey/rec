package main

import (
	"fmt"
	"github.com/williamsharkey/rec/portaudio"
	"github.com/williamsharkey/tui-go-copy"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	//miditest()
	recsNameSize := loadRecs()

	err := portaudio.Initialize()
	if err != nil {
		return
	}
	defer portaudio.Terminate()

	//recs := loadRecs()
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

	recToggle := func() {
		if rs.Rec.Active {
			rs.CmdList.ChangeItem("REC", "rec")
		} else {
			rs.CmdList.ChangeItem("rec", "REC")
		}
		go clickRec(rs)
	}

	cmdList := tui.NewList()
	rs.CmdList = cmdList
	cmdList.AddItems(
		//"noise",
		"rec",
		//"nums",
		"midi",
		"exit")
	cmdList.Select(0)
	cmdList.OnItemActivated(func(l *tui.List) {
		switch strings.ToLower(l.SelectedItem()) {
		case "midi":
			playnote(1, 60)
		case "noise":
			go noise()
		case "exit":
			exit(ui)
		case "rec":
			recToggle()
		case "nums":
			go clickNum(rs)
		default:
		}
	})

	cmdBox := tui.NewVBox(cmdList)
	cmdBox.SetBorder(true)

	histList := tui.NewList()
	histListBox := tui.NewVBox(histList)
	histListBox.SetBorder(true)
	histListBox.SetTitle("log")

	recList := tui.NewList()
	recListBox := tui.NewVBox(recList)
	recList.OnItemActivated(func(l *tui.List) {
		go play(l.SelectedItem())
	})
	recList.OnSelectionChanged(func(l *tui.List) {

		recListBox.SetTitle("wavs - " + l.SelectedItem() + fmt.Sprintf("- %d Bytes", rs.SizeList[l.Selected()].Size))
	})

	rs.RecList = recList

	recAppend(rs, recsNameSize...)

	recListBox.SetBorder(true)
	recListBox.SetTitle("wavs")

	input := tui.NewEntry()
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetTitle("input")

	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(recListBox, histListBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)

	hbox.Append(cmdBox)
	hbox.Append(chat)
	tui.DefaultFocusChain.Set(cmdList, recList, input)

	ui.SetKeybinding("Esc", func() { exit(ui) })
	ui.SetKeybinding("R", recToggle)

	input.OnSubmit(func(e *tui.Entry) {
		t := e.Text()
		if t == "quit" || t == "exit" {
			exit(ui)
			return
		}
		recAppend(rs, nameSize{e.Text(), 0})
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

func loadRecs() (recsNameSize []nameSize) {
	path := filepath.Join(".", "wavs")
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}
	for _, file := range files {
		n := file.Name()
		ext := filepath.Ext(n)
		name := strings.TrimSuffix(n, ext)
		if err != nil || strings.ToLower(ext) != ".wav" {
			continue
		}

		recsNameSize = append(recsNameSize, nameSize{name, file.Size()})

	}
	return
}

type RecSettings struct {
	Rec, Play *AudioChan
	RecSlice  []int32
	PlaySlice []int16
	RecList   *tui.List
	SizeList  []nameSize
	UI        *tui.UI
	CmdList   *tui.List
}
type AudioChan struct {
	Active   bool
	Kill     chan int
	Complete chan int
	Buffer   chan [1024]int32
	Print    chan string
	PAStream *portaudio.Stream
}

func recAppend(rs *RecSettings, fnsize ...nameSize) {
	s := []string{}
	for _, x := range fnsize {
		s = append(s, x.Name)
	}
	rs.RecList.AddItems(s...)
	rs.SizeList = append(rs.SizeList, fnsize...)
	rs.RecList.Select(rs.RecList.Length() - 1)
	(*rs.UI).Repaint()
}
