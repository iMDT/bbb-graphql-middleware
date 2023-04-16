package common

import "context"

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
}
