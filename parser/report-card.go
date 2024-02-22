package parser

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
)

type ModuleReport struct {
	Identifier   string         `json:"id"`
	Name         string         `json:"name"`
	Year         uint           `json:"year"`
	PassingGrade string         `json:"passingGrade"`
	GlobalGrade  string         `json:"grade"`
	Credits      uint           `json:"credits"`
	Situation    string         `json:"situation"`
	Classes      []*ModuleClass `json:"classes"`
}

type ModuleClass struct {
	Identifier string        `json:"id"`
	Name       string        `json:"name"`
	Grades     []*ClassGrade `json:"grades"`
	Mean       string        `json:"mean"`
	Weight     uint          `json:"weight"`
}

type ClassGrade struct {
	Name   string `json:"name"`
	Weight uint   `json:"weight"`
	Grade  string `json:"grade"`
}

type reportCardRowType int
type reportCardParser struct {
	Parser

	doc *goquery.Document
}

const (
	reportCardTableHeader reportCardRowType = iota
	reportCardModuleRow
	reportCardUnitRow
	reportCardCreditsRow
	reportCardUnknownRow = -1
)

var (
	UnknownReportCardStructure = errors.New("unknown report card structure")
)

func (s *Parser) ReportCard() ([]*ModuleReport, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s.src))
	if err != nil {
		return nil, err
	}

	return reportCardParser{
		Parser: *s,
		doc:    doc,
	}.parse()
}

func (p reportCardParser) parse() ([]*ModuleReport, error) {
	var reports []*ModuleReport
	var globalErr error

	p.doc.Find("table#record_table tr").Each(func(i int, s *goquery.Selection) {
		if globalErr != nil {
			return
		}

		switch p.getRowType(s) {
		case reportCardTableHeader:
			if s.Children().Length() < 6 {
				globalErr = UnknownReportCardStructure
				return
			}
		case reportCardModuleRow:
			module, err := p.parseModuleRow(s)
			if err != nil {
				globalErr = err
				return
			}

			reports = append(reports, module)
		case reportCardUnitRow:
			class, err := p.parseUnitRow(s)
			if err != nil {
				globalErr = err
				return
			}

			moduleOff := len(reports) - 1
			reports[moduleOff].Classes = append(reports[moduleOff].Classes, class)
		}
	})

	return reports, globalErr
}

func (p reportCardParser) getRowType(row *goquery.Selection) reportCardRowType {
	if row.HasClass("bulletin_header_row") {
		return reportCardTableHeader
	} else if row.HasClass("bulletin_module_row") {
		if row.HasClass("total-credits-row") {
			return reportCardCreditsRow
		}
		return reportCardModuleRow
	} else if row.HasClass("bulletin_unit_row") {
		return reportCardUnitRow
	} else {
		return reportCardUnknownRow
	}
}

func (p reportCardParser) parseModuleRow(row *goquery.Selection) (*ModuleReport, error) {
	id := row.Find("td.module-code").Text()

	nameText := row.Find("td").Eq(1).Text()
	nameSplit := strings.SplitN(nameText, id, 2)
	name := strings.TrimSpace(nameSplit[0][:len(nameSplit[0])-1])
	passingGrade := strings.TrimSpace(nameSplit[1][len(") [seuil : ") : len(nameSplit[1])-1])

	situation := row.Find("td").Eq(2).Text()

	year, _ := strconv.ParseUint(strings.SplitN(row.Find("td").Eq(3).Text(), " - ", 2)[0], 10, 32)

	grade := row.Find("td").Eq(4).Text()

	credits, err := strconv.ParseUint(row.Find("td").Eq(6).Text(), 10, 32)
	if err != nil {
		return nil, err
	}

	return &ModuleReport{
		Identifier:   id,
		Name:         name,
		Year:         uint(year),
		PassingGrade: passingGrade,
		GlobalGrade:  grade,
		Credits:      uint(credits),
		Situation:    situation,
	}, nil
}

func (p reportCardParser) parseUnitRow(row *goquery.Selection) (*ModuleClass, error) {
	id := row.Find("td").Eq(0).Text()

	classContents := row.Find("td").Eq(1).Contents()
	className := strings.TrimSpace(classContents.First().Text())

	var grades []*ClassGrade
	for i := 2; i < classContents.Length(); i += 2 {
		if strings.TrimSpace(classContents.Eq(i).Text()) == "" {
			break
		}

		gradeText := strings.SplitN(classContents.Eq(i).Text(), "(", 2)
		name := strings.TrimSpace(gradeText[0])
		weight, _ := strconv.ParseUint(strings.TrimSpace(gradeText[1][:strings.Index(gradeText[1], "%")]), 10, 32)
		grade := strings.TrimSpace(classContents.Eq(i + 1).Text())

		grades = append(grades, &ClassGrade{
			Name:   name,
			Weight: uint(weight),
			Grade:  grade,
		})
	}

	mean := row.Find("td").Eq(4).Text()

	weight, _ := strconv.ParseUint(row.Find("td").Eq(5).Text(), 10, 32)

	return &ModuleClass{
		Identifier: id,
		Name:       className,
		Grades:     grades,
		Mean:       mean,
		Weight:     uint(weight),
	}, nil
}
