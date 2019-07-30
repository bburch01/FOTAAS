package actions

func (as *ActionSuite) Test_TelemetryHandler() {
	res := as.HTML("/telemetry").Get()

	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Telemetry Page")
}
