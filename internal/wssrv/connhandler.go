package wssrv

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/iMDT/bbb-graphql-middleware/internal/hascli"
	"nhooyr.io/websocket"
)

var lastConnectionId int

// Buffer size of the channels
var bufferSize = 100

// Handle client connection
// This is the connection that comes from browser
func WebsocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// Obtain id for this connection
	lastConnectionId++
	connectionId := fmt.Sprintf("%010d", lastConnectionId)

	// Starts a context that will be dependent on the connection, so we can cancel subroutines when the connection is dropped
	browserConnectionContext, browserConnectionContextCancel := context.WithCancel(r.Context())
	defer browserConnectionContextCancel()

	// Add sub-protocol
	var acceptOptions websocket.AcceptOptions
	acceptOptions.Subprotocols = append(acceptOptions.Subprotocols, "graphql-ws")

	c, err := websocket.Accept(w, r, &acceptOptions)
	if err != nil {
		log.Printf("[%v WebsocketConnectionHandler] error: %v", connectionId, err)
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// Log it
	log.Printf("[%v WebsocketConnectionHandler] connection accepted", connectionId)

	// Create channels
	fromBrowserChannel := make(chan interface{}, bufferSize)
	toBrowserChannel := make(chan interface{}, bufferSize)

	go hascli.HasuraClient(connectionId, browserConnectionContext, r.Cookies(), fromBrowserChannel, toBrowserChannel)

	// Configure the wait group (to hold this routine execution until both are completed)
	var wg sync.WaitGroup
	wg.Add(3)

	// Start routines
	// reads from browser, writes to fromBrowserMirrorChannel
	var fromBrowserReplicatedChannel = make(chan interface{}, bufferSize)
	go SessionTokenReader(connectionId, browserConnectionContext, fromBrowserReplicatedChannel, fromBrowserChannel, &wg)

	go WebsocketConnectionReader(connectionId, browserConnectionContext, c, fromBrowserReplicatedChannel, toBrowserChannel, &wg)
	go WebsocketConnectionWriter(connectionId, browserConnectionContext, c, fromBrowserReplicatedChannel, toBrowserChannel, &wg)

	// Wait
	wg.Wait()

}
