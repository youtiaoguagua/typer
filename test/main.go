package main

import (
	"github.com/schollz/progressbar/v3"
	"time"
)

func main() {
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSpinnerType(29),
	)
	for i := 0; i < 1000; i++ {
		bar.Add(1)
		time.Sleep(5 * time.Millisecond)
	}
}
