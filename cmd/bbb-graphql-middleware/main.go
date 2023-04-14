package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/iMDT/bbb-graphql-middleware/internal/wssrv"
)

func main() {
	var listenPort = 8378
	http.HandleFunc("/", wssrv.WebsocketConnectionHandler)
	log.Printf("[main] listening on port %v", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", listenPort), nil))
}
