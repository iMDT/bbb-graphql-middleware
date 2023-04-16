package hascli

func HasuraRestoreState(hc *HasuraConnection, fromBrowserChannel chan interface{}) {
	// for query :
	for _, query := range hc.browserconn.CurrentQueries {
		if query.LastSeenOnHasuraConnetion != hc.id {
			fromBrowserChannel <- query.Message
		}
	}
}
