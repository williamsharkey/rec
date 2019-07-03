package main

import (
	"fmt"
	"github.com/williamsharkey/rec/go-wave"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
)

var sig = make(chan os.Signal, 1)

func init() {

	signal.Notify(sig, os.Interrupt, os.Kill)
}
func clickRec(rs *RecSettings) {
	s := rs.Rec
	s.Active = !s.Active

	if rs.Rec.Active {

		go recNew(rs)

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
		defer waveFile.Close()

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
		defer waveWriter.Close()
	Loop2:
		for {
			select {
			case <-sig:
				return
				//panic("yo")
				//waveWriter.Close()
				//waveFile.Close()
				//histAppend(rs.RecList, rs.UI, fn)
				//break Loop2
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
				//if err==nil{waveWriter

				histAppend(rs, nameSize{fn, int64(waveWriter.RiffChunk.Size)}) //f.Size()})
				//}
				break Loop2

			default:
			}
		}
	} else {
		rs.Rec.Kill <- 1
	}
}
