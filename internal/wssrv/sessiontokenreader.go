package wssrv

import (
	"context"
	"log"
	"sync"
)

func SessionTokenReader(connectionId string, browserConnectionContext context.Context, input chan interface{}, output chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Printf("[%v SessionTokenReader] finished", connectionId)

	// Consume the fromBrowser channel and pass it

	var sessionToken = "N.A."

	// Intercept the fromBrowserMessage channel to get the sessionToken
	for fromBrowserMessage := range input {
		// Sends to the output channel (to be used by other routines)
		output <- fromBrowserMessage

		// Gets the sessionToken
		if sessionToken == "N.A." {
			var fromBrowserMessageAsMap = fromBrowserMessage.(map[string]interface{})

			if fromBrowserMessageAsMap["type"] == "connection_init" {
				var payloadAsMap = fromBrowserMessageAsMap["payload"].(map[string]interface{})
				var headersAsMap = payloadAsMap["headers"].(map[string]interface{})
				sessionToken := headersAsMap["X-Session-Token"].(string)
				log.Printf("[%v SessionTokenReader] intercepted session token %v", connectionId, sessionToken)
				WsConnections[connectionId].SessionToken = sessionToken
			}
		}
	}
}
