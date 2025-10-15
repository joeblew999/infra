package ui

import "strings"

const overviewRoute = "overview"

// NormalizePage returns the page route that should be rendered, falling back to
// the overview when an unknown route is requested.
func NormalizePage(snapshot Snapshot, requested string) string {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return overviewRoute
	}
	if requested == overviewRoute {
		return overviewRoute
	}
	if snapshot.ServiceDetails != nil {
		if _, ok := snapshot.ServiceDetails[requested]; ok {
			return requested
		}
	}
	return overviewRoute
}
