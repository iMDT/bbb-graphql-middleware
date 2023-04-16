package hascli

import (
	"log"
	"sync"

	"github.com/iMDT/bbb-graphql-middleware/internal/common"
	"nhooyr.io/websocket/wsjson"
)

// HasuraConnectionWriter
// process messages (middleware->hasura)
func HasuraConnectionWriter(hc *common.HasuraConnection, fromBrowserChannel chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer hc.ContextCancelFunc()
	defer log.Printf("[%v HasuraConnectionWriter] finished", hc.Browserconn.Id)

RangeLoop:
	for {
		select {
		case <-hc.Context.Done():
			break RangeLoop
		case fromBrowserMessage := <-fromBrowserChannel:
			{
				var fromBrowserMessageAsMap = fromBrowserMessage.(map[string]interface{})

				if fromBrowserMessageAsMap["type"] == "start" {
					var queryId = fromBrowserMessageAsMap["id"].(string)
					hc.Browserconn.CurrentQueries[queryId] = common.GraphQlQuery{
						Id:                        queryId,
						Message:                   fromBrowserMessage,
						LastSeenOnHasuraConnetion: hc.Id,
					}

					log.Printf("[%v %v HasuraConnectionWriter] Current queries: %v", hc.Browserconn.Id, hc.Id, hc.Browserconn.CurrentQueries, hc)
				}

				if fromBrowserMessageAsMap["type"] == "stop" {
					var queryId = fromBrowserMessageAsMap["id"].(string)
					delete(hc.Browserconn.CurrentQueries, queryId)

					log.Printf("[%v %v HasuraConnectionWriter] Current queries: %v", hc.Browserconn.Id, hc.Id, hc.Browserconn.CurrentQueries)
				}

				if fromBrowserMessageAsMap["type"] == "connection_init" {
					hc.Browserconn.ConnectionInitMessage = fromBrowserMessage
				}

				log.Printf("[%v %v HasuraConnectionWriter] [middleware->hasura] %v", hc.Browserconn.Id, hc.Id, fromBrowserMessage)
				err := wsjson.Write(hc.Context, hc.Websocket, fromBrowserMessage)
				if err != nil {
					log.Printf("[%v %v HasuraConnectionWriter] error on write (we're disconnected from hasura): %v", hc.Browserconn.Id, hc.Id, err)
					return
				}
			}
		}
	}
}
