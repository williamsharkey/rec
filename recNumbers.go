package main

func recNumbers(rs *RecSettings) (err error) {
	x := [1024]int32{1, 2, 3, 0, -1, -2, -3}
	s := rs.Rec
	rs.Rec.Print <- "started recording numbers"

	s.Buffer <- x //[]int32{1,2,3,-1,-2,-3}
	s.Complete <- 1
	s.Print <- "complete rec numbers "
	return
}
