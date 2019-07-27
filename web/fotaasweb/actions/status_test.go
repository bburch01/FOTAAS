package actions

func (as *ActionSuite) Test_StatusHandler() {
	res := as.HTML("/status").Get()

	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Status Page")
}
