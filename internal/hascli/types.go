package hascli

import (
	"context"

	"github.com/iMDT/bbb-graphql-middleware/internal/common"
	"nhooyr.io/websocket"
)

// hascli
type HasuraConnection struct {
	id                string                    // hasura connection id
	browserconn       *common.BrowserConnection // browser connection that originated this hasura connection
	websocket         *websocket.Conn           // websocket used to connect to hasura
	context           context.Context           // hasura connection context
	contextCancelFunc context.CancelFunc        // function to cancel the hasura context (and so, the connection)
}
