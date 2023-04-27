package main

import (
	"fmt"
	"github.com/iMDT/bbb-graphql-middleware/internal/wssrv"
	"github.com/iMDT/bbb-graphql-middleware/internal/wssrv/invalidator"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	// Configure logger
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log := log.WithField("_routine", "SessionTokenReader")

	// Connection invalidator
	go invalidator.RedisConnectionnInvalidator()

	// Websocket listener
	var listenPort = 8378
	http.HandleFunc("/", wssrv.WebsocketConnectionHandler)

	log.Infof("listening on port %v", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", listenPort), nil))

}
