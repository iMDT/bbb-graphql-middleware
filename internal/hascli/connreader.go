package hascli

import (
	"log"
	"sync"

	"nhooyr.io/websocket/wsjson"
)

// HasuraConnectionReader
// process messages (hasura->middleware)
// toBrowserChannel - channel that this routine continuously read
// fromBrowserChannel - channel that this routine write the previous queries on conn_ack
func HasuraConnectionReader(hc *HasuraConnection, toBrowserChannel chan interface{}, fromBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer hc.contextCancelFunc()
	defer log.Printf("[%v HasuraConnectionReader] finished", hc.browserconn.Id)

	for {
		var message interface{}
		err := wsjson.Read(hc.context, hc.websocket, &message)
		if err != nil {
			return
		}

		log.Printf("[%v HasuraConnectionReader] [hasura->middleware] %v", hc.browserconn.Id, message)

		toBrowserChannel <- message

		var messageAsMap = message.(map[string]interface{})

		if messageAsMap["type"] == "connection_ack" {
			// writes previous queries to hasura
			go HasuraRestoreState(hc, fromBrowserChannel)
		}

	}
}
