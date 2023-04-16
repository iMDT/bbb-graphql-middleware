package hascli

import (
	"log"
	"sync"

	"github.com/iMDT/bbb-graphql-middleware/internal/common"
	"nhooyr.io/websocket/wsjson"
)

// HasuraConnectionWriter
// process messages (middleware->hasura)
func HasuraConnectionWriter(hc *HasuraConnection, fromBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer hc.contextCancelFunc()
	defer log.Printf("[%v HasuraConnectionWriter] finished", hc.browserconn.Id)

RangeLoop:
	for {
		select {
		case <-hc.context.Done():
			break RangeLoop
		case fromBrowserMessage := <-fromBrowserChannel:
			{
				var fromBrowserMessageAsMap = fromBrowserMessage.(map[string]interface{})

				if fromBrowserMessageAsMap["type"] == "start" {
					var queryId = fromBrowserMessageAsMap["id"].(string)
					hc.browserconn.CurrentQueries[queryId] = common.GraphQlQuery{
						Id:                        queryId,
						Message:                   fromBrowserMessage,
						LastSeenOnHasuraConnetion: hc.id,
					}

					log.Printf("[%v HasuraConnectionWriter] Current queries: %v", hc.browserconn.Id, hc.browserconn.CurrentQueries)
				}

				if fromBrowserMessageAsMap["type"] == "stop" {
					var queryId = fromBrowserMessageAsMap["id"].(string)
					delete(hc.browserconn.CurrentQueries, queryId)

					log.Printf("[%v HasuraConnectionWriter] Current queries: %v", hc.browserconn.Id, hc.browserconn.CurrentQueries)
				}

				if fromBrowserMessageAsMap["type"] == "connection_init" {
					hc.browserconn.ConnectionInitMessage = fromBrowserMessage
				}

				log.Printf("[%v HasuraConnectionWriter] [middleware->hasura] %v", hc.browserconn.Id, fromBrowserMessage)
				err := wsjson.Write(hc.context, hc.websocket, fromBrowserMessage)
				if err != nil {
					log.Printf("[%v HasuraConnectionWriter] error on write (we're disconnected from hasura): %v", hc.browserconn.Id, err)
					return
				}
			}
		}
	}
}
