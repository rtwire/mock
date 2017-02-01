package main

import (
	"flag"
	"fmt"
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

	url := fmt.Sprintf("http://localhost:%d/v1/mainnet/", *port)
	log.Printf("RTWire service running at %s.", url)

	addr := ":" + strconv.Itoa(*port)
	log.Fatal(http.ListenAndServe(addr, service.New()))
}
