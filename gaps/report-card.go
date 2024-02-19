package gaps

import (
	"fmt"
	"golang.org/x/net/html/charset"
	"lutonite.dev/gaps-cli/parser"
)

type ReportCardAction struct {
	cfg *TokenClientConfiguration
}

func NewReportCardAction(config *TokenClientConfiguration) *ReportCardAction {
	return &ReportCardAction{
		cfg: config,
	}
}

func (a *ReportCardAction) FetchReportCard() ([]*parser.ModuleReport, error) {
	req, err := a.cfg.buildRequest("GET", fmt.Sprintf("/consultation/notes/bulletin.php?id=%d", a.cfg.studentId))
	if err != nil {
		return nil, err
	}

	res, err := a.cfg.doForm(req, nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	utfBody, err := charset.NewReader(res.Body, "iso-8859-1")
	if err != nil {
		return nil, err
	}

	pres, err := parser.FromResponseBody(utfBody)
	if err != nil {
		return nil, err
	}

	return pres.ReportCard()
}
