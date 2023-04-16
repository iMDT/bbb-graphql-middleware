package hascli

import "github.com/iMDT/bbb-graphql-middleware/internal/common"

func HasuraRestoreState(hc *common.HasuraConnection, fromBrowserChannel chan interface{}) {
	// for query :
	for _, query := range hc.Browserconn.CurrentQueries {
		if query.LastSeenOnHasuraConnetion != hc.Id {
			fromBrowserChannel <- query.Message
		}
	}
}
