package wssrv

import (
	"context"
	log "github.com/sirupsen/logrus"
	"sync"
)

func SessionTokenReader(connectionId string, browserConnectionContext context.Context, input chan interface{}, wg *sync.WaitGroup) {
	log := log.WithField("_routine", "SessionTokenReader")

	defer wg.Done()
	defer log.Info("finished")

	// Consume the fromBrowser channel and pass it

	var sessionToken = "N.A."

	// Intercept the fromBrowserMessage channel to get the sessionToken
	for fromBrowserMessage := range input {
		// Gets the sessionToken
		if sessionToken == "N.A." {
			var fromBrowserMessageAsMap = fromBrowserMessage.(map[string]interface{})

			if fromBrowserMessageAsMap["type"] == "connection_init" {
				var payloadAsMap = fromBrowserMessageAsMap["payload"].(map[string]interface{})
				var headersAsMap = payloadAsMap["headers"].(map[string]interface{})
				var sessionToken = headersAsMap["X-Session-Token"]
				if sessionToken != nil {
					sessionToken := headersAsMap["X-Session-Token"].(string)
					log.Printf("[%v SessionTokenReader] intercepted session token %v", connectionId, sessionToken)
					WsConnections[connectionId].SessionToken = sessionToken
				}
			}
		}
	}
}
