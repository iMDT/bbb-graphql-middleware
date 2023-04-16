package hascli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"github.com/iMDT/bbb-graphql-middleware/internal/common"
	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
)

var lastHasuraConnectionId int
var hasuraEndpoint = "ws://127.0.0.1:8080/v1/graphql"
var bufferSize = 10

// Hasura client connection
func HasuraClient(browserConnection *common.BrowserConnection, cookies []*http.Cookie, fromBrowserChannel chan interface{}, toBrowserChannel chan interface{}) error {
	defer log.Printf("[%v hasuraClient] finished", browserConnection.Id)

	// Obtain id for this connection
	lastHasuraConnectionId++
	hasuraConnectionId := "HC" + fmt.Sprintf("%010d", lastHasuraConnectionId)

	// Add sub-protocol
	var dialOptions websocket.DialOptions
	dialOptions.Subprotocols = append(dialOptions.Subprotocols, "graphql-ws")

	// Create cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return xerrors.Errorf("failed to create cookie jar: %w", err)
	}
	parsedURL, err := url.Parse(hasuraEndpoint)
	if err != nil {
		return xerrors.Errorf("failed to parse url: %w", err)
	}
	parsedURL.Scheme = "http"
	jar.SetCookies(parsedURL, cookies)
	hc := &http.Client{
		Jar: jar,
	}
	dialOptions.HTTPClient = hc

	// Create a context for the hasura connection, that depends on the browser context
	// this means that if browser connection is closed, the hasura connection will close also
	// this also means that we can close the hasura connection without closing the browser one
	// this is used for the invalidation process (reconnection only on the hasura side )
	var hasuraConnectionContext, hasuraConnectionContextCancel = context.WithCancel(browserConnection.Context)
	defer hasuraConnectionContextCancel()

	var thisConnection = HasuraConnection{
		id:                hasuraConnectionId,
		browserconn:       browserConnection,
		context:           hasuraConnectionContext,
		contextCancelFunc: hasuraConnectionContextCancel,
	}

	// Make the connection
	c, _, err := websocket.Dial(hasuraConnectionContext, hasuraEndpoint, &dialOptions)
	if err != nil {
		return xerrors.Errorf("error connecting to hasura: %v", err)
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	thisConnection.websocket = c

	// Log the connection success
	log.Printf("[%v hasuraClient] connected with Hasura", browserConnection.Id)

	// Configure the wait group
	var wg sync.WaitGroup
	wg.Add(2)

	// Start routines

	// reads from browser, writes to hasura
	go HasuraConnectionWriter(&thisConnection, fromBrowserChannel, &wg)

	// reads from hasura, writes to browser
	go HasuraConnectionReader(&thisConnection, toBrowserChannel, fromBrowserChannel, &wg)

	// if it's a reconnect, inject authentication
	if thisConnection.browserconn.ConnectionInitMessage != nil {
		fromBrowserChannel <- thisConnection.browserconn.ConnectionInitMessage
	}

	// Wait
	wg.Wait()

	return nil
}
