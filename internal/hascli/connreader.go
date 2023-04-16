package hascli

import (
	"log"
	"sync"

	"github.com/iMDT/bbb-graphql-middleware/internal/common"
	"nhooyr.io/websocket/wsjson"
)

// HasuraConnectionReader
// process messages (hasura->middleware)
// toBrowserChannel - channel that this routine continuously read
// fromBrowserChannel - channel that this routine write the previous queries on conn_ack
func HasuraConnectionReader(hc *common.HasuraConnection, toBrowserChannel chan interface{}, fromBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer hc.ContextCancelFunc()
	defer log.Printf("[%v %v HasuraConnectionReader] finished", hc.Browserconn.Id, hc.Id)

	for {
		var message interface{}
		err := wsjson.Read(hc.Context, hc.Websocket, &message)
		if err != nil {
			return
		}

		log.Printf("[%v %v HasuraConnectionReader] [hasura->middleware] %v", hc.Browserconn.Id, hc.Id, message)

		toBrowserChannel <- message

		var messageAsMap = message.(map[string]interface{})

		if messageAsMap["type"] == "connection_ack" {
			// writes previous queries to hasura
			go HasuraRestoreState(hc, fromBrowserChannel)
		}

	}
}
