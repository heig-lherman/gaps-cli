package gaps

import (
	"fmt"
	"net/url"

	"lutonite.dev/gaps-cli/parser"
)

type AbsencesAction struct {
	cfg  *TokenClientConfiguration
	year uint
}

func NewAbsencesAction(config *TokenClientConfiguration, year uint) *AbsencesAction {
	return &AbsencesAction{
		cfg:  config,
		year: year,
	}
}

func (a *AbsencesAction) FetchAbsences() (*parser.AbsenceReport, error) {
	req, err := a.cfg.buildRequest("POST", "/consultation/etudiant/")
	if err != nil {
		return nil, err
	}

	// POST rsargs to get all absences
	showAllConfig := fmt.Sprintf(`["studentAbsGrid_rateSelectorId","studentAbsGrid","%s",null,null,"%d","0",%d,null]`,
		a.cfg.token, a.year, a.cfg.studentId)

	data := url.Values{}
	data.Add("rs", "smartReplacePart")
	data.Add("rsargs", showAllConfig)

	res, err := a.cfg.doForm(req, data)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	pres, err := parser.FromResponseBody(res.Body)
	if err != nil {
		return nil, err
	}

	absences, err := pres.Absences()
	if err != nil {
		return nil, err
	}
	return absences, nil

}
