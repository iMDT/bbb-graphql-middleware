package hascli

import (
	"context"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
)

var hasuraEndpoint = "ws://127.0.0.1:8080/v1/graphql"

// var hasuraEndpoint = "ws://127.0.0.1:8888/v1/graphql"
var hasuraConnectionContexts map[string]context.Context
var bufferSize = 10

// Hasura client connection
func HasuraClient(connectionId string, browserConnectionContext context.Context, cookies []*http.Cookie, fromBrowserChannel chan interface{}, toBrowserChannel chan interface{}) error {
	defer log.Printf("[%v hasuraClient] finished", connectionId)

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
	var hasuraConnectionContext, hasuraConnectionContextCancel = context.WithCancel(browserConnectionContext)
	defer hasuraConnectionContextCancel()

	// Make the connection
	c, _, err := websocket.Dial(hasuraConnectionContext, hasuraEndpoint, &dialOptions)
	if err != nil {
		return xerrors.Errorf("error connecting to hasura: %v", err)
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// Log the connection success
	log.Printf("[%v hasuraClient] connected with Hasura", connectionId)

	// Configure the wait group
	var wg sync.WaitGroup
	wg.Add(2)

	// Start routines

	// reads from browser, writes to hasura
	go HasuraConnectionWriter(connectionId, hasuraConnectionContext, hasuraConnectionContextCancel, c, fromBrowserChannel, &wg)

	// reads from hasura, writes to browser
	go HasuraConnectionReader(connectionId, hasuraConnectionContext, hasuraConnectionContextCancel, c, toBrowserChannel, &wg)

	// Wait
	wg.Wait()

	return nil
}
