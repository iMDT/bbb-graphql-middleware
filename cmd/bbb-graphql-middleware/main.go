package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Buffer size of the channels
var bufferSize = 100
var lastConnectionId int
var hasuraEndpoint = "ws://127.0.0.1:8080/v1/graphql"
var listenPort = 8378

// hasuraClientReader
// process messages (hasura->middleware)
func hasuraClientReader(connectionId string, ctx context.Context, c *websocket.Conn, toBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Printf("[%v hasuraClientReader] finished", connectionId)

	for {
		var message interface{}
		err := wsjson.Read(ctx, c, &message)
		if err != nil {
			return
		}

		log.Printf("[%v hasuraClientReader] [hasura->middleware] %v", connectionId, message)

		toBrowserChannel <- message
	}
}

// hasuraClientWriter
// process messages (middleware->hasura)
func hasuraClientWriter(connectionId string, ctx context.Context, c *websocket.Conn, fromBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Printf("[%v hasuraClientWriter] finished", connectionId)

	for fromBrowserMessage := range fromBrowserChannel {
		log.Printf("[%v hasuraClientWriter] [middleware->hasura] %v", connectionId, fromBrowserMessage)
		err := wsjson.Write(ctx, c, fromBrowserMessage)
		if err != nil {
			return
		}
	}
}

// Hasura client connection
func hasuraClient(connectionId string, ctx context.Context, cookies []*http.Cookie, fromBrowserChannel chan interface{}, toBrowserChannel chan interface{}) error {
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

	// Make the connection
	c, _, err := websocket.Dial(ctx, hasuraEndpoint, &dialOptions)
	if err != nil {
		return xerrors.Errorf("error connecting to hasura: %v", err)
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// Log the connection success
	log.Printf("[%v hasuraClient] connected with Hasura", connectionId)

	// Configure the wait group (to hold this routine execution until both are completed)
	var wg sync.WaitGroup
	wg.Add(2)

	// Start routines
	go hasuraClientWriter(connectionId, ctx, c, fromBrowserChannel, &wg)
	go hasuraClientReader(connectionId, ctx, c, toBrowserChannel, &wg)

	// Wait
	wg.Wait()

	return nil
}

func websocketConnectionReader(connectionId string, ctx context.Context, c *websocket.Conn, fromBrowserChannel chan interface{}, toBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(fromBrowserChannel)
	defer close(toBrowserChannel)
	defer log.Printf("[%v websocketConnectionReader] finished", connectionId)
	for {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var v interface{}
		err := wsjson.Read(ctx, c, &v)
		if err != nil {
			log.Printf("[%v websocketHandler] error: %v", connectionId, err)
			return
		}

		log.Printf("[%v websocketHandler] [browser->middleware] received: %v", connectionId, v)

		fromBrowserChannel <- v
	}
}

func websocketConnectionWriter(connectionId string, ctx context.Context, c *websocket.Conn, fromBrowserChannel chan interface{}, toBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Printf("[%v websocketConnectionWriter] finished", connectionId)

	for toBrowserMessage := range toBrowserChannel {
		// If no messages could be sent to browser after 30s, disconnect (keep alives goes server->client)
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()

		log.Printf("[%v websocketConnectionWriter] [middleware->browser] %v", connectionId, toBrowserMessage)
		err := wsjson.Write(ctx, c, toBrowserMessage)
		if err != nil {
			return
		}
	}
}

// Handle client connection
// This is the connection that comes from browser
func websocketConnection(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("[%v websocketHandler] error: %v", connectionId, err)
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// Log it
	log.Printf("[%v websocketHandler] connection accepted", connectionId)

	// Create channels
	fromBrowserChannel := make(chan interface{}, bufferSize)
	toBrowserChannel := make(chan interface{}, bufferSize)

	go hasuraClient(connectionId, browserConnectionContext, r.Cookies(), fromBrowserChannel, toBrowserChannel)

	// Configure the wait group (to hold this routine execution until both are completed)
	var wg sync.WaitGroup
	wg.Add(2)

	// Start routines
	go websocketConnectionReader(connectionId, browserConnectionContext, c, fromBrowserChannel, toBrowserChannel, &wg)
	go websocketConnectionWriter(connectionId, browserConnectionContext, c, fromBrowserChannel, toBrowserChannel, &wg)

	// Wait
	wg.Wait()

}

func main() {
	http.HandleFunc("/", websocketConnection)
	log.Printf("[main] listening on port %v", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", listenPort), nil))
}
