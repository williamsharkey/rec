package main

import (
	"fmt"
)

func recNew(rs *RecSettings) (err error) {

	recArr := [1024]int32{}
	s := rs.Rec
	err = s.PAStream.Start()
	if err != nil {
		return
	}
	rs.Rec.Print <- "started recording"

	for {

		select {

		case x := <-s.Kill:
			s.PAStream.Stop()
			//if err != nil {
			//	return
			//}
			s.Print <- fmt.Sprintf("killed with %d", x)

			s.Complete <- 1
			return
		//case <-sig:
		//	s.PAStream.Stop()
		//	//if err != nil {
		//	//	return
		//	//}
		//	s.Print <- "killed with os sig"
		//
		//	s.Complete <- 1
		//	break recloop

		default:
		}

		err = s.PAStream.Read()
		if err != nil {
			s.PAStream.Stop()
			//if err != nil {
			//	return
			//}
			s.Print <- "rec err " + err.Error()
			s.Complete <- 1

			return
		}
		copy(recArr[:], rs.RecSlice[:])
		s.Buffer <- recArr

	}
	return
}
