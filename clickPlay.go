package main

import "fmt"

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
