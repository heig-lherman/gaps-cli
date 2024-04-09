package gaps

import (
	ics "github.com/arran4/golang-ical"
)

func GetAllClasses(cfg *TokenClientConfiguration, year uint) []string {
	sa0 := NewStudentScheduleAction(cfg, year, 0)
	sa1 := NewStudentScheduleAction(cfg, year, 1)
	sa2 := NewStudentScheduleAction(cfg, year, 3)
	s0, _ := sa0.FetchSchedule()
	s1, _ := sa1.FetchSchedule()
	s2, _ := sa2.FetchSchedule()

	classes := make([]string, 0)

	if s0 != nil {
		for _, event := range s0.Events() {
			classes = append(classes, event.GetProperty(ics.ComponentPropertySummary).Value)
		}
	}

	if s1 != nil {
		for _, event := range s1.Events() {
			classes = append(classes, event.GetProperty(ics.ComponentPropertySummary).Value)
		}
	}

	if s2 != nil {
		for _, event := range s2.Events() {
			classes = append(classes, event.GetProperty(ics.ComponentPropertySummary).Value)
		}
	}

	return classes
}
