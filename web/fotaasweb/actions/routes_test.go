package actions

func (as *ActionSuite) Test_RoutesHandler() {
	res := as.HTML("/routes").Get()

	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Routes Page")
}
