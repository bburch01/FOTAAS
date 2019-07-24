package grifts

import (
	"github.com/bburch01/FOTAAS/web/fotaasweb/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
