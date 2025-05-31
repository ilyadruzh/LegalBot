package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("listen", ":8080", "listen address")
	flag.Parse()

	log.Printf("starting bot on %s", *addr)
	if err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})); err != nil {
		log.Fatal(err)
	}
}
