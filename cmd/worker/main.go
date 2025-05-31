package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	queue := flag.String("queue", "amqp://guest:guest@localhost:5672/", "AMQP URL")
	flag.Parse()

	log.Printf("starting worker, queue=%s", *queue)
	for range time.Tick(time.Second) {
		log.Println("worker tick")
	}
}
