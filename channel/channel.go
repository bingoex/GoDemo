package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

var jobs int

func init() {
	flag.IntVar(&jobs, "job", 1000, "jobs to run cocorently")
	flag.Parse()
}

func work(result chan int) {
	n := rand.Intn(5 * 1000)
	time.Sleep(time.Duration(n) * time.Microsecond)
	result <- n
}

func main() {
	resultChan := make(chan int, jobs)

	for i := 0; i < jobs; i++ {
		go work(resultChan)
	}

	totalCost := 0
	for i := 0; i < jobs; i++ {
		totalCost += <-resultChan
	}

	fmt.Printf("%d workers total cost:%d\n", jobs, totalCost)
}
