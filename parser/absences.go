package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type AbsenceReport struct {
	Student     string
	Orientation string
	Courses     []CourseAbsence
}

type CourseAbsence struct {
	Name    string
	Periods struct {
		Ete   int
		Term1 int
		Term2 int
		Term3 int
		Term4 int
	}
	Total           int
	Justified       int
	RelativePeriods int
	AbsolutePeriods int
}

func (s *Parser) Absences() (*AbsenceReport, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s.src))
	if err != nil {
		return nil, err
	}
	report := &AbsenceReport{}

	report.Student = doc.Find(".s_cell").First().Text()
	report.Orientation = doc.Find(".l_cell.s_cell").First().Text()

	doc.Find("tr.a_r_0").Each(func(i int, row *goquery.Selection) {
		if row.Find("td.l_cell").Length() == 0 {
			return
		}

		course := CourseAbsence{}

		courseName := row.Find("td.l_cell:not(.s_cell)").First().Text()
		course.Name = strings.TrimSpace(courseName)

		cells := row.Find("td.b_cell")

		course.Periods.Ete = parseAbsenceWithJustified(cells.Eq(0).Text(), &course.Justified)
		course.Periods.Term1 = parseAbsenceWithJustified(cells.Eq(1).Text(), &course.Justified)
		course.Periods.Term2 = parseAbsenceWithJustified(cells.Eq(2).Text(), &course.Justified)
		course.Periods.Term3 = parseAbsenceWithJustified(cells.Eq(3).Text(), &course.Justified)
		course.Periods.Term4 = parseAbsenceWithJustified(cells.Eq(4).Text(), &course.Justified)
		course.Total = parseAbsence(cells.Eq(5).Text())

		course.RelativePeriods = parseAbsence(cells.Eq(6).Text())
		course.AbsolutePeriods = parseAbsence(cells.Eq(7).Text())

		report.Courses = append(report.Courses, course)
	})

	return report, nil
}

func parseAbsenceWithJustified(text string, totalJustified *int) int {
	text = strings.TrimSpace(text)
	if text == "" || text == "&nbsp" {
		return 0
	}

	// Look for pattern like "2 [2]"
	if matches := regexp.MustCompile(`(\d+)\s*\[(\d+)\]`).FindStringSubmatch(text); matches != nil {
		justified, _ := strconv.Atoi(matches[2])
		*totalJustified += justified // Add to total justified
		total, _ := strconv.Atoi(matches[1])
		return total
	}

	num, _ := strconv.Atoi(text)
	return num
}

func parseAbsence(text string) int {
	text = strings.TrimSpace(text)
	if text == "" || text == "&nbsp" {
		return 0
	}
	// Remove any [x] if present and just get the main number
	if idx := strings.Index(text, "["); idx != -1 {
		text = strings.TrimSpace(text[:idx])
	}
	num, _ := strconv.Atoi(text)
	return num
}
