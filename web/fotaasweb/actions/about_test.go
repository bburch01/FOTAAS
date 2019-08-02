package actions

func (as *ActionSuite) Test_AboutHandler() {
	res := as.HTML("/about").Get()

	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "About Page")
}
