package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/rtwire/mock/service"
)

var (
	port = flag.Int("port", 8085, "service port number")
)

func main() {
	flag.Parse()

	addr := ":" + strconv.Itoa(*port)

	log.Printf("Mock RTWire service running on port %d.", *port)
	log.Fatal(http.ListenAndServe(addr, service.New()))
}
