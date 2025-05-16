package main

import (
	"fmt"
)

func stage1() <-chan int {
	out := make(chan int)
	go func() {
		for i := 1; i <= 5; i++ {
			out <- i
		}
		close(out)
	}()
	return out
}

func stage2(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for v := range in {
			out <- v * 2
		}
		close(out)
	}()
	return out
}

func stage3(in <-chan int) <-chan string {
	out := make(chan string)
	go func() {
		for v := range in {
			out <- fmt.Sprintf("Value: %d", v)
		}
		close(out)
	}()
	return out
}

func main() {
	c := stage3(stage2(stage1())) // the pipeline
	for msg := range c {
		fmt.Println(msg)
	}
}
