package actions

import "github.com/gobuffalo/buffalo"

// AnalysisHandler is a default handler to serve up
// the analysis page.
func AnalysisHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("analysis.html"))
}
