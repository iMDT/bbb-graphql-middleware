package wssrv

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

func RedisConnectionnInvalidator() {
	var ctx = context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	subscriber := redisClient.Subscribe(ctx, "from-akka-apps-redis-channel")

	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("[RedisConnectionnInvalidator] error: ", err)
		}

		var message interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			panic(err)
		}

		messageAsMap := message.(map[string]interface{})

		messageEnvelopeAsMap := messageAsMap["envelope"].(map[string]interface{})

		messageType := messageEnvelopeAsMap["name"]

		if messageType == "InvalidateUserGraphqlConnectionSysMsg" {
			messageCoreAsMap := messageAsMap["core"].(map[string]interface{})
			messageBodyAsMap := messageCoreAsMap["body"].(map[string]interface{})
			sessionTokenToInvalidate := messageBodyAsMap["sessionToken"]
			log.Printf("[RedisConnectionnInvalidator] Received invalidate request for sessionToken %v", sessionTokenToInvalidate)

			for _, browserConnection := range WsConnections {
				if browserConnection.SessionToken == sessionTokenToInvalidate {
					if browserConnection.HasuraConnection != nil {
						log.Printf("[RedisConnectionnInvalidator] Processing invalidate request for sessionToken %v (hasura connection %v)", sessionTokenToInvalidate, browserConnection.HasuraConnection.Id)
						browserConnection.HasuraConnection.ContextCancelFunc()
						log.Printf("[RedisConnectionnInvalidator] Processed invalidate request for sessionToken %v (hasura connection %v)", sessionTokenToInvalidate, browserConnection.HasuraConnection.Id)
					}
				}
			}
		}
	}
}
