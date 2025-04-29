package gaps

import (
	"fmt"
	"net/url"

	"lutonite.dev/gaps-cli/parser"
)

type RegistryAction struct {
	cfg  *TokenClientConfiguration
	year uint
}

func NewRegistryAction(config *TokenClientConfiguration, year uint) *RegistryAction {
	return &RegistryAction{
		cfg:  config,
		year: year,
	}
}

func (r *RegistryAction) FetchRegistry() (*parser.Registry, error) {
	req, err := r.cfg.buildRequest("POST", "/consultation/horaires/")
	if err != nil {
		return nil, err
	}

	// POST rsargs to get all sections
	showAllConfig := fmt.Sprintf(`[%d,2,false,{"0":{"6":{"27":true,"28":true,"29":true,"30":true,"31":true}},"1":{"8":{"8":true,"16":true,"18":true},"9":{"9":true,"32":true,"33":true,"34":true},"10":{"36":true,"37":true,"48":true},"12":{"53":true},"15":{"24":true,"54":true}},"2":{"2":{"35":true,"38":true,"40":true},"14":{"14":true,"49":true,"50":true,"51":true}},"3":{"13":{"13":true}},"4":{"4":{"4":true}},"16":{"2":{"35":true},"10":{"48":true},"15":{"24":true}},"17":{"3":{"3":true}},"63":{"7":{"7":true}},"-1":true},null]`, r.year)

	data := url.Values{}
	data.Add("rs", "getMenuHoraire")
	data.Add("rsargs", showAllConfig)

	res, err := r.cfg.doForm(req, data)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	pres, err := parser.FromResponseBody(res.Body)
	if err != nil {
		return nil, err
	}

	registry, err := pres.Registry()
	if err != nil {
		return nil, err
	}

	return registry, nil
}
