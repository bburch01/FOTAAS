package actions

import "github.com/gobuffalo/buffalo"

// TelemetryHandler is a default handler to serve up
// the telemetry page.
func TelemetryHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("telemetry.html"))
}
