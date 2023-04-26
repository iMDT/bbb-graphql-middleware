package reader

import (
	"context"
	log "github.com/sirupsen/logrus"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func WebsocketConnectionReader(browserConnectionId string, ctx context.Context, c *websocket.Conn, fromBrowserChannel1 chan interface{}, fromBrowserChannel2 chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(fromBrowserChannel1)
	log := log.WithField("_routine", "WebsocketConnectionReader").WithField("browserConnectionId", browserConnectionId)
	defer log.Infof("finished")
	for {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var v interface{}
		err := wsjson.Read(ctx, c, &v)
		if err != nil {
			log.Errorf("error on read (browser is disconnected): %v", err)
			return
		}

		log.Tracef("received from browser: %v", v)

		fromBrowserChannel1 <- v
		fromBrowserChannel2 <- v
	}
}
