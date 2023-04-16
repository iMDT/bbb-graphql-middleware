package common

import (
	"context"

	"nhooyr.io/websocket"
)

// wssrv
type GraphQlQuery struct {
	Id                        string
	Message                   interface{}
	LastSeenOnHasuraConnetion string // id of the hasura connection that this query was active
}

type BrowserConnection struct {
	Id                    string
	SessionToken          string
	Context               context.Context
	CurrentQueries        map[string]GraphQlQuery
	ConnectionInitMessage interface{}
	HasuraConnection      *HasuraConnection
}

type HasuraConnection struct {
	Id                string             // hasura connection id
	Browserconn       *BrowserConnection // browser connection that originated this hasura connection
	Websocket         *websocket.Conn    // websocket used to connect to hasura
	Context           context.Context    // hasura connection context
	ContextCancelFunc context.CancelFunc // function to cancel the hasura context (and so, the connection)
}
