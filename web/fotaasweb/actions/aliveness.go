package actions

import "github.com/gobuffalo/buffalo"

// AlivenessHandler is a default handler to serve up
// the aliveness page.
func AlivenessHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("aliveness.html"))
}
