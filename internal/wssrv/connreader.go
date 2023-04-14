package wssrv

import (
	"context"
	"log"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func WebsocketConnectionReader(connectionId string, ctx context.Context, c *websocket.Conn, fromBrowserChannel chan interface{}, toBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(fromBrowserChannel)
	defer log.Printf("[%v websocketConnectionReader] finished", connectionId)
	for {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var v interface{}
		err := wsjson.Read(ctx, c, &v)
		if err != nil {
			log.Printf("[%v WebsocketConnectionReader] error on read (browser is disconnected): %v", connectionId, err)
			return
		}

		log.Printf("[%v WebsocketConnectionReader] [browser->middleware] received: %v", connectionId, v)

		fromBrowserChannel <- v
	}
}
