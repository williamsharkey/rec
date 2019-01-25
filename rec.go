package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/zenwerk/go-wave"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"unsafe"

	//"math"
	"os"
	"os/signal"
	"strconv"
	//"unsafe"
)

func errCheck(err error) {

	if err != nil {
		panic(err)
	}
}

func copyArr(a []int16) (b []int16) {
	b = make([]int16, len(a))
	copy(b, a)
	return b
}

func flatten(slices [][]int16) (flattened []int16) {

	for _, s := range slices {
		flattened = append(flattened, s...)
	}
	return flattened
}

func slice(a []int16) (slices [][]int16) {
	lastValue := int16(0)
	j := 0
	for i, v := range a {
		if lastValue < 0 && v >= 0 {
			slices = append(slices, a[j:i])
			j = i
		}
		lastValue = v
	}

	if j != len(a)-1 {
		slices = append(slices, a[j:])
	}
	return slices
}

func sortSlices(slices [][]int16) {

	sort.Slice(slices, func(i, j int) bool {
		return len(slices[i]) < len(slices[j])
	})

}

func sortSlices2(slices [][]int16) {
	sort.Slice(slices, func(i, j int) bool {
		return slices[i][0] < slices[j][0]
	})
}

func effect(clean []int16) (effect []int16) {
	effect = copyArr(clean)
	s := slice(effect)
	sortSlices(s)
	f := flatten(s)
	return f
}

func blur(clean []int16) (effect []int16) {
	effect = copyArr(clean)
	cs := slice(effect)
	lenc := len(cs)
	cs = append(cs, cs[lenc-1])
	for i := 0; i < lenc; i++ {
		lc := len(cs[i])
		r1 := resamp(cs[i+1], lc)
		cs[i] = comb(cs[i], r1)
	}

	f := flatten(cs[0:lenc])
	return f
}

func blur2(clean []int16) (effect []int16) {
	effect = copyArr(clean)
	cs := slice(effect)
	lenc := len(cs)
	cs = append(cs, cs[lenc-1])
	for i := 0; i < lenc; i++ {
		lc := len(cs[i])
		r1 := resamp(cs[i+1], lc)
		cs[i] = xfade(cs[i], r1)
	}

	f := flatten(cs[0:lenc])
	return f
}

func blur2n(clean []int16, n int) []int16 {

	b := clean
	for i := 0; i < n; i++ {
		b = blur2(b)
	}
	return b

}

func comb(a []int16, b []int16) (both []int16) {

	lena := len(a)
	both = make([]int16, lena, lena)
	for i := 0; i < lena; i++ {
		both[i] = int16(math.Round(float64(a[i]+b[i]) / 2))
	}
	return
}

func xfade(a []int16, b []int16) (fade []int16) {

	lena := len(a)
	fade = make([]int16, lena, lena)
	for i := 0; i < lena; i++ {
		aProp := float64(i) / float64(lena-1)
		bProp := 1.0 - aProp
		fade[i] = int16(math.Round(float64(a[i])*aProp + float64(b[i])*bProp))
	}
	return
}

func mul(a []int16, b []int16) (both []int16) {

	lena := len(a)
	both = make([]int16, lena, lena)
	for i := 0; i < lena; i++ {
		both[i] = int16(math.Round(float64(a[i])*float64(b[i])) / (math.MaxInt16 / 4))
	}
	return
}

func inv(a []int16) (ef []int16) {

	lena := len(a)
	ef = make([]int16, lena, lena)
	for i := 0; i < lena; i++ {
		ef[i] = -a[i]
	}
	return
}

func effect2(clean []int16) (effect []int16) {
	effect = copyArr(clean)
	s := slice(effect)
	sortSlices2(s)
	f := flatten(s)
	return f
}

func resamp(a []int16, m int) (r []int16) {
	n := len(a)
	r = make([]int16, m, m)
	a = append(a, a[len(a)-1])
	for i := 0; i < m; i++ {

		f := float64(i*n) / float64(m)
		ff := math.Floor(f)
		fFloor := int(ff)
		//fmt.Printf("%d %f %d\n",i,f,fFloor)

		curr := a[fFloor]
		next := a[fFloor+1]
		mixNext := f - ff
		mixCurrent := float64(1) - mixNext
		r[i] = int16(math.Round(mixCurrent*float64(curr) + mixNext*float64(next)))
	}
	return r
}

func convUnsafe(b []byte) []int16 {
	return (*(*[]int16)(unsafe.Pointer(&b)))[0 : len(b)/2]
}

func convSafe(b []byte) []int16 {
	r := make([]int16, len(b)/2, len(b)/2)

	for i := 0; i < len(b)/2; i++ {
		r[i] = int16(b[i*2+1])<<8 + int16(b[i*2])
	}

	return r
}

func read(name string) (audio []int16) {

	r, err := wave.NewReader(name + ".wav")
	errCheck(err)
	b := make([]byte, r.NumSamples*2, r.NumSamples*2)
	r.Read(b)

	audio = convSafe(b)

	return
}
func readImg(path string) image.Image {
	infile, err := os.Open(path)
	if err != nil {
		// replace this with real error handling
		panic(err.Error())
	}
	defer infile.Close()
	src, _, err := image.Decode(infile)
	if err != nil {
		// replace this with real error handling
		panic(err.Error())
	}
	return src
}
func imgToWaveTable(img image.Image) (int16s []int16) {

	// Create a new grayscale image
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			oldColor := img.At(x, y)

			g := color.Gray16Model.Convert(oldColor).(color.Gray16)
			int16s = append(int16s, int16(int(g.Y)-math.MaxUint16+math.MaxInt16))

		}
	}
	return
}

func imgToWaveTableFixed(img image.Image) (int16s []int16) {

	// Create a new grayscale image
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	for x := 0; x < 256; x++ {
		for y := 0; y < 2048; y++ {
			oldColor := img.At(x*w/256, y*h/2048)

			g := color.Gray16Model.Convert(oldColor).(color.Gray16)
			int16s = append(int16s, int16(int(g.Y)-math.MaxUint16+math.MaxInt16))

		}
	}
	return
}
func imgToWaveTableFixed90(img image.Image) (int16s []int16) {

	// Create a new grayscale image
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	for y := 0; y < 256; y++ {
		for x := 0; x < 2048; x++ {

			oldColor := img.At(x*w/2048, y*h/256)

			g := color.Gray16Model.Convert(oldColor).(color.Gray16)
			int16s = append(int16s, int16(int(g.Y)-math.MaxUint16+math.MaxInt16))

		}
	}
	return
}

func main() {
	//t:=[]int16{0, 2, 4}
	//z:=resamp(t,6)
	//fmt.Printf("%+v\n",z )
	//fmt.Printf("%+v\n",t )
	//return
	//clean := read("test")
	//wet := effect(clean)
	//write("test"+".wet", wet)

	//write("test", []int16{1, 2, 3})
	//read("test")
	//return

	fmt.Printf("%+v\n", os.Args[1:])

	if len(os.Args) > 3 {

		if os.Args[1] == "rec" {

			cutoff := 1

			sec, err := strconv.ParseInt(os.Args[3], 10, 32)
			if err == nil {
				cutoff = int(sec)
			}
			name := os.Args[2]
			fmt.Printf("rec command, filename: %s, seconds %d", name, cutoff)
			rec(cutoff, name)
			return
		}

	}

	if len(os.Args) > 2 {
		if os.Args[1] == "wt" {
			img := readImg(os.Args[2])
			wt := imgToWaveTableFixed(img)
			write(os.Args[2]+".wt", wt)
			return
		}
		if os.Args[1] == "slice" {
			clean := read(os.Args[2])
			wet := effect(clean)
			write(os.Args[2]+".slice", wet)
			return
		}

		if os.Args[1] == "slice2" {
			clean := read(os.Args[2])
			wet := effect2(clean)
			write(os.Args[2]+".slice2", wet)
			return
		}

		if os.Args[1] == "blur" {
			clean := read(os.Args[2])
			wet := blur(clean)
			write(os.Args[2]+".blur", wet)
			return
		}
		if os.Args[1] == "blur2" {
			clean := read(os.Args[2])
			wet := blur2n(clean, 10)
			write(os.Args[2]+".blur2", wet)
			return
		}

	}
	if len(os.Args) > 1 {
		if os.Args[1] == "wtall" {
			filepath.Walk(".", func(path string, f os.FileInfo, _ error) error {
				if f.IsDir() {
					return nil
				}
				if strings.Contains(path, ".git/") {
					return nil
				}
				if strings.Contains(path, "/") {
					return nil
				}
				if filepath.Ext(path) == ".png" {
					fmt.Println("process " + path)
					img := readImg(path)
					pathT := strings.TrimRight(path, ".png")
					wt := imgToWaveTableFixed(img)

					write(pathT, wt)
					wt = imgToWaveTableFixed90(img)

					write(pathT+"_rot", wt)
				}
				return nil
			})

			return
		}
	}
}

func write(s string, int16s []int16) {
	waveFile, err := os.Create(s + ".wav")
	param := wave.WriterParam{
		Out:           waveFile,
		Channel:       1,
		SampleRate:    44100,
		BitsPerSample: 16, // if 16, change to WriteSample16()
	}

	waveWriter, err := wave.NewWriter(param)
	errCheck(err)
	_, err = waveWriter.WriteSample16(int16s)

	errCheck(err)

	waveWriter.Close()

}

func rec(cutoff int, name string) {
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt, os.Kill)

	audioFileName := name + ".wav" //os.Args[1]

	fmt.Println("Recording to "+audioFileName, cutoff, "seconds")

	waveFile, err := os.Create(audioFileName)
	errCheck(err)

	inputChannels := 1
	outputChannels := 0
	sampleRate := 44100
	buf := 4410
	int16Slice := make([]int16, buf)

	portaudio.Initialize()
	//defer portaudio.Terminate()

	apis, err := portaudio.HostApis()

	for i, api := range apis {

		fmt.Println(i, api.Name, api.Type)
		for j, d := range api.Devices {

			fmt.Println(" ", j, d.Name, d.MaxInputChannels, d.MaxOutputChannels)
		}
	}

	stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, float64(sampleRate), len(int16Slice), int16Slice)
	errCheck(err)
	//defer stream.Close()

	// setup Wave file writer

	param := wave.WriterParam{
		Out:           waveFile,
		Channel:       inputChannels,
		SampleRate:    sampleRate,
		BitsPerSample: 16, // if 16, change to WriteSample16()
	}

	waveWriter, err := wave.NewWriter(param)
	errCheck(err)

	defer waveWriter.Close()

	// start reading from microphone
	errCheck(stream.Start())
	fmt.Println("Recording is live now. Say something to your microphone!")
	i := 0

	//uInt16Slice := make([]uint16, buf, buf)
	for {

		errCheck(stream.Read())

		i++

		//uInt16Slice = *(*[]uint16)(unsafe.Pointer(&int16Slice))
		_, err := waveWriter.WriteSample16(int16Slice)

		errCheck(err)
		elapsed := (i * buf) / sampleRate
		fmt.Printf("\r %d", elapsed)
		if cutoff != 0 && elapsed >= cutoff {
			close(waveWriter, stream)
		}
		select {
		case <-sig:
			close(waveWriter, stream)
		default:
		}
	}

	errCheck(stream.Stop())
}

func close(waveWriter *wave.Writer, stream *portaudio.Stream) {
	fmt.Println("\nclosed")

	waveWriter.Close()
	stream.Close()
	portaudio.Terminate()
	os.Exit(0)
}
