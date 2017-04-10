package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/rtwire/mock/service"
)

var (
	addr = flag.String("addr", ":8085", "service address")
)

func main() {
	flag.Parse()

	url := fmt.Sprintf("http://%s/v1/mainnet/", *addr)
	log.Printf("RTWire service running at %s.", url)

	log.Fatal(http.ListenAndServe(*addr, service.New()))
}
