package actions

import "github.com/gobuffalo/buffalo"

// StatusHandler is a default handler to serve up
// a status page.
func StatusHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("status.html"))
}
