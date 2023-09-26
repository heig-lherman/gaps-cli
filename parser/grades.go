package parser

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ClassGrades struct {
	Name        string        `json:"name"`
	GlobalMean  string        `json:"globalMean"`
	HasExam     bool          `json:"hasExam"`
	GradeGroups []*GradeGroup `json:"gradeGroups"`
}

type GradeGroup struct {
	Name   string   `json:"name"`
	Mean   string   `json:"mean"`
	Weight uint     `json:"weight"`
	Grades []*Grade `json:"grades"`
}

type Grade struct {
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Weight      float32   `json:"weight"`
	Grade       string    `json:"grade"`
	ClassMean   string    `json:"classMean"`
}

type gradeRowType int
type gradesParser struct {
	Parser

	doc *goquery.Document
}

const (
	classHeader gradeRowType = iota
	groupHeader
	gradeRow
	unknownRow
)

func (s *Parser) Grades() ([]*ClassGrades, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s.src))
	if err != nil {
		return nil, err
	}

	return gradesParser{
		Parser: *s,
		doc:    doc,
	}.parse()
}

func (p gradesParser) parse() ([]*ClassGrades, error) {
	var classes []*ClassGrades
	var globalErr error
	p.doc.Find("table.displayArray tbody tr").Each(func(i int, s *goquery.Selection) {
		switch p.getRowType(s) {
		case classHeader:
			class, err := p.parseClassHeader(s)
			if err != nil {
				globalErr = err
				return
			}

			classes = append(classes, class)
		case groupHeader:
			group, err := p.parseGroupHeader(s)
			if err != nil {
				globalErr = err
				return
			}

			classOff := len(classes) - 1

			classes[classOff].GradeGroups = append(classes[classOff].GradeGroups, group)
		case gradeRow:
			grade, err := p.parseGradeRow(s)
			if err != nil {
				globalErr = err
				return
			}

			classOff := len(classes) - 1
			groupOff := len(classes[classOff].GradeGroups) - 1

			if groupOff < 0 {
				return
			}

			classes[classOff].GradeGroups[groupOff].Grades = append(
				classes[classOff].GradeGroups[groupOff].Grades,
				grade,
			)
		default:
			globalErr = errors.New("unknown row type")
			log.Debugf("unknown row with content:\n%s", s.Text())
		}
	})

	return classes, globalErr
}

func (p gradesParser) getRowType(row *goquery.Selection) gradeRowType {
	if row.Has("td.bigheader").Length() > 0 {
		return classHeader
	} else if row.Has("td[rowspan]").Length() > 0 {
		return groupHeader
	} else if row.Find("td").Size() == 5 {
		return gradeRow
	} else {
		return unknownRow
	}
}

func (p gradesParser) parseClassHeader(row *goquery.Selection) (*ClassGrades, error) {
	text := row.Find("td.bigheader").Text()

	re, err := regexp.Compile(`(?m)^([\w-]{3,}) - moyenne\s+?(hors examen)?.+(\d.\d+|-)$`)
	if err != nil {
		return nil, err
	}

	matches := re.FindStringSubmatch(text)
	if len(matches) != 4 {
		return nil, fmt.Errorf("could not parse class header")
	}

	return &ClassGrades{
		Name:       matches[1],
		GlobalMean: matches[3],
		HasExam:    matches[2] != "",
	}, nil
}

func (p gradesParser) parseGroupHeader(row *goquery.Selection) (*GradeGroup, error) {
	text, err := row.Find("td").First().Html()
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(`(?m)^(.+)<br\s*?/?>moyenne : (\d.\d|-)<br\s*?/?>poids : (\d+)$`)
	if err != nil {
		return nil, err
	}

	matches := re.FindStringSubmatch(text)
	if len(matches) != 4 {
		return nil, fmt.Errorf("could not parse class header")
	}

	weight, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return nil, err
	}

	return &GradeGroup{
		Name:   matches[1],
		Mean:   matches[2],
		Weight: uint(weight),
	}, nil
}

func (p gradesParser) parseGradeRow(row *goquery.Selection) (*Grade, error) {
	tds := row.Find("td")
	dateCell := tds.Eq(0) // date fmt dd.mm.yyyy
	descriptionCell := tds.Eq(1)
	meanCell := tds.Eq(2)
	weightCell := tds.Eq(3)
	gradeCell := tds.Eq(4)

	date, err := time.Parse("02.01.2006", dateCell.Text())
	if err != nil {
		return nil, err
	}

	weight, err := p.parseWeight(weightCell)
	if err != nil {
		return nil, err
	}

	return &Grade{
		Description: p.parseDescription(descriptionCell),
		Date:        date,
		Weight:      weight,
		Grade:       gradeCell.Text(),
		ClassMean:   meanCell.Text(),
	}, nil
}

func (p gradesParser) parseWeight(weightCell *goquery.Selection) (float32, error) {
	re, err := regexp.Compile(`(?m)^.+?\(([\d.]+)%\)$`)
	if err != nil {
		return 0, err
	}

	matches := re.FindStringSubmatch(weightCell.Text())
	if len(matches) != 2 {
		return 0, fmt.Errorf("could not parse weight")
	}

	weight, err := strconv.ParseFloat(matches[1], 32)
	if err != nil {
		return 0, err
	}

	return float32(weight), nil
}

func (p gradesParser) parseDescription(descriptionCell *goquery.Selection) string {
	if descriptionCell.Has("div[onclick]").Length() > 0 {
		return strings.TrimSpace(descriptionCell.Find("div div").Last().Text())
	} else {
		return descriptionCell.Text()
	}
}
