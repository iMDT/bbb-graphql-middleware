package hascli

import (
	"context"
	"log"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// HasuraConnectionWriter
// process messages (middleware->hasura)
func HasuraConnectionWriter(connectionId string, hasuraConnectionContext context.Context, hasuraConnectionContextCancel context.CancelFunc, c *websocket.Conn, fromBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer hasuraConnectionContextCancel()
	defer log.Printf("[%v HasuraConnectionWriter] finished", connectionId)

RangeLoop:
	for {
		select {
		case <-hasuraConnectionContext.Done():
			break RangeLoop
		case fromBrowserMessage := <-fromBrowserChannel:
			{
				log.Printf("[%v HasuraConnectionWriter] [middleware->hasura] %v", connectionId, fromBrowserMessage)
				err := wsjson.Write(hasuraConnectionContext, c, fromBrowserMessage)
				if err != nil {
					log.Printf("[%v HasuraConnectionWriter] error on write (we're disconnected from hasura): %v", connectionId, err)
					return
				}
			}
		}
	}
}
