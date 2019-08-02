package actions

func (as *ActionSuite) Test_AnalysisHandler() {
	res := as.HTML("/analysis").Get()

	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Analysis Page")
}