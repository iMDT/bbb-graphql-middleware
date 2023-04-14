package hascli

import (
	"context"
	"log"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// HasuraConnectionReader
// process messages (hasura->middleware)
func HasuraConnectionReader(connectionId string, hasuraConnectionContext context.Context, hasuraConnectionContextCancel context.CancelFunc, c *websocket.Conn, toBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer hasuraConnectionContextCancel()
	defer log.Printf("[%v HasuraConnectionReader] finished", connectionId)

	for {
		var message interface{}
		err := wsjson.Read(hasuraConnectionContext, c, &message)
		if err != nil {
			return
		}

		log.Printf("[%v HasuraConnectionReader] [hasura->middleware] %v", connectionId, message)

		toBrowserChannel <- message
	}
}
