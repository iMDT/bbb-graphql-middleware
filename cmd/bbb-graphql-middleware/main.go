package main

import (
	"fmt"
	"github.com/iMDT/bbb-graphql-middleware/internal/wssrv"
	"github.com/iMDT/bbb-graphql-middleware/internal/wssrv/invalidator"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	// Define the log level (TraceLevel logs everything)
	log.SetLevel(log.TraceLevel)
	// Define log format as JSON
	log.SetFormatter(&log.JSONFormatter{})
	// Specify the routine that emitted the log
	log := log.WithField("_routine", "SessionTokenReader")

	// Connection invalidator
	go invalidator.RedisConnectionnInvalidator()

	// Webscoket listener
	var listenPort = 8378
	http.HandleFunc("/", wssrv.WebsocketConnectionHandler)

	log.Infof("listening on port %v", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", listenPort), nil))

}
