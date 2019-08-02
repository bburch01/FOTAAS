package actions

import "github.com/gobuffalo/buffalo"

// SimulationHandler is a default handler to serve up
// a similation page.
func SimulationHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("simulation.html"))
}
