package main

import (
	"log"

	"github.com/cainmusic/goworker"
)

func main() {
	square()
	//errorForAddTask()
}

func square() {
	pool, err := goworker.NewPool(20)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= 500; i++ {
		n := i
		pool.AddTask(func() {
			log.Print(n, n*n)
		})
	}

	pool.Done()
}

func errorForAddTask() {
	pool, _ := goworker.NewPool(5)
	pool.Done()

	f := func() {
		log.Print("hi")
	}

	if err := pool.AddTask(f); err != nil {
		log.Fatal(err) // task queue done, no more task
	}
}
