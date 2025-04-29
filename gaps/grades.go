package gaps

import (
	"errors"
	"fmt"
	"lutonite.dev/gaps-cli/parser"
	"net/url"
)

type GradesAction struct {
	cfg *TokenClientConfiguration

	year     uint
	semester Semester

	ClassFilter string
}

func NewGradesAction(config *TokenClientConfiguration, year uint) *GradesAction {
	return NewSemesterGradesAction(config, year, All)
}

func NewSemesterGradesAction(config *TokenClientConfiguration, year uint, semester Semester) *GradesAction {
	return &GradesAction{
		cfg:      config,
		year:     year,
		semester: semester,
	}
}

func (a *GradesAction) FetchGrades() ([]*parser.ClassGrades, error) {
	req, err := a.cfg.buildRequest("POST", "/consultation/controlescontinus/consultation.php")
	if err != nil {
		return nil, err
	}

	res, err := a.cfg.doForm(req, url.Values{
		"rs":     {"getStudentCCs"},
		"rsargs": {fmt.Sprintf("[%d, %d, %d]", a.cfg.studentId, a.year, a.semester.rsArg())},
	})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	pres, err := parser.FromResponseBody(res.Body)
	if err != nil {
		return nil, err
	}

	classes, err := pres.Grades()
	if err != nil {
		return nil, err
	}

	if a.ClassFilter != "" {
		return a.findClass(classes)
	}

	return classes, nil
}

// filter classes by name
func (a *GradesAction) findClass(classes []*parser.ClassGrades) ([]*parser.ClassGrades, error) {
	var filtered []*parser.ClassGrades
	for _, class := range classes {
		if class.Name == a.ClassFilter {
			filtered = append(filtered, class)
		}
	}

	if len(filtered) == 0 {
		return nil, errors.New("no class found with name " + a.ClassFilter)
	}

	return filtered, nil
}
