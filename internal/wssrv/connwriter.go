package wssrv

import (
	"context"
	"log"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func WebsocketConnectionWriter(connectionId string, ctx context.Context, c *websocket.Conn, fromBrowserChannel chan interface{}, toBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Printf("[%v websocketConnectionWriter] finished", connectionId)

RangeLoop:
	for {
		select {
		case <-ctx.Done():
			break RangeLoop
		case toBrowserMessage := <-toBrowserChannel:
			{
				log.Printf("[%v websocketConnectionWriter] [middleware->browser] %v", connectionId, toBrowserMessage)
				err := wsjson.Write(ctx, c, toBrowserMessage)
				if err != nil {
					log.Printf("[%v websocketConnectionWriter] error on write (browser is disconnected): %v", connectionId, err)
					return
				}
			}
		}
	}
}
