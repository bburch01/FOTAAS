package actions

func (as *ActionSuite) Test_SimulationHandler() {
	res := as.HTML("/simulation").Get()

	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Simulation Page")
}
