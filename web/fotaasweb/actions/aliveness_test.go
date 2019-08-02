package actions

func (as *ActionSuite) Test_AlivenessHandler() {
	res := as.HTML("/aliveness").Get()

	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Aliveness Page")
}
