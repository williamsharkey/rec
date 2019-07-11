package main

import (
	"fmt"
	"github.com/williamsharkey/go-wave"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type nameSize struct {
	Name string
	Size int64
}

func clickNum(rs *RecSettings) {
	//s := rs.Rec
	//s.Active = !s.Active

	//if rs.Rec.Active {
	go recNumbers(rs)

	newpath := filepath.Join(".", "wavs")
	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		return
	}
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
		BitsPerSample: 32,
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
			//x:=[4*1024]byte{}
			//for i,s:=range samps{
			//	p:=int32ToByte(s)
			//	x[i+0]=p[0]
			//	x[i+1]=p[1]
			//	x[i+2]=p[2]
			//	x[i+3]=p[3]
			//}
			_, err = waveWriter.WriteSample32(samps[:])

			if err != nil {
				waveWriter.Close()
				waveFile.Close()
				return
			}

		case <-rs.Rec.Complete:
			waveWriter.Close()
			waveFile.Close()
			//f,err:=waveFile.Stat()
			//if err==nil{
			recAppend(rs, nameSize{fn, int64(waveWriter.RiffChunk.Size + 8)}) //f.Size()})
			//}

			break Loop2
		default:
		}
	}
	//} else {
	//	rs.Rec.Kill <- 1
	//}
}
