package main

import (
	"fmt"
	"time"
)

var or func(channels ...<-chan any) <-chan any

func init() {
	or = func(channels ...<-chan any) <-chan any {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}

		out := make(chan any)
		go func() {
			defer close(out)
			mid := len(channels) / 2
			select {
			case <-or(channels[:mid]...):
			case <-or(channels[mid:]...):
			}
		}()
		return out
	}
}

func main() {
	sig := func(after time.Duration) <-chan any {
		c := make(chan any)
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))
}
