package wssrv

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/iMDT/bbb-graphql-middleware/internal/common"
	"github.com/iMDT/bbb-graphql-middleware/internal/hascli"
	"nhooyr.io/websocket"
)

var lastBrowserConnectionId int

// Buffer size of the channels
var bufferSize = 100

// active websocket connections
var WsConnections map[string]*common.BrowserConnection = make(map[string]*common.BrowserConnection)

// Handle client connection
// This is the connection that comes from browser
func WebsocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// Obtain id for this connection
	lastBrowserConnectionId++
	browserConnectionId := "BC" + fmt.Sprintf("%010d", lastBrowserConnectionId)

	// Starts a context that will be dependent on the connection, so we can cancel subroutines when the connection is dropped
	browserConnectionContext, browserConnectionContextCancel := context.WithCancel(r.Context())
	defer browserConnectionContextCancel()

	// Add sub-protocol
	var acceptOptions websocket.AcceptOptions
	acceptOptions.Subprotocols = append(acceptOptions.Subprotocols, "graphql-ws")

	c, err := websocket.Accept(w, r, &acceptOptions)
	if err != nil {
		log.Printf("[%v WebsocketConnectionHandler] error: %v", browserConnectionId, err)
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	var thisConnection = common.BrowserConnection{
		Id:             browserConnectionId,
		CurrentQueries: make(map[string]common.GraphQlQuery, 1),
		Context:        browserConnectionContext,
	}

	WsConnections[browserConnectionId] = &thisConnection

	defer delete(WsConnections, browserConnectionId)

	// Log it
	log.Printf("[%v WebsocketConnectionHandler] connection accepted", browserConnectionId)

	// Create channels
	fromBrowserChannel := make(chan interface{}, bufferSize)
	toBrowserChannel := make(chan interface{}, bufferSize)

	// Ensure a hasura client is running while the browser is connected
	go func() {
		log.Printf("[%v WebsocketConnectionHandler] starting hasura client", browserConnectionId)

	BrowserConnectedLoop:
		for {
			select {
			case <-browserConnectionContext.Done():
				break BrowserConnectedLoop
			default:
				{
					log.Printf("[%v WebsocketConnectionHandler] creating hasura client", browserConnectionId)
					hascli.HasuraClient(WsConnections[browserConnectionId], r.Cookies(), fromBrowserChannel, toBrowserChannel)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()

	// Configure the wait group (to hold this routine execution until both are completed)
	var wg sync.WaitGroup
	wg.Add(3)

	// Start routines
	// reads from browser, writes to fromBrowserMirrorChannel
	var fromBrowserReplicatedChannel = make(chan interface{}, bufferSize)
	go SessionTokenReader(browserConnectionId, browserConnectionContext, fromBrowserReplicatedChannel, fromBrowserChannel, &wg)

	go WebsocketConnectionReader(browserConnectionId, browserConnectionContext, c, fromBrowserReplicatedChannel, toBrowserChannel, &wg)
	go WebsocketConnectionWriter(browserConnectionId, browserConnectionContext, c, fromBrowserReplicatedChannel, toBrowserChannel, &wg)

	// Wait
	wg.Wait()

}
