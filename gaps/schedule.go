package gaps

import (
	"fmt"
	"github.com/arran4/golang-ical"
)

type ScheduleAction struct {
	cfg       *TokenClientConfiguration
	year      uint
	semester  uint
	schedType uint
	targetId  uint
}

func NewStudentScheduleAction(config *TokenClientConfiguration, year uint, semester uint) *ScheduleAction {
	return &ScheduleAction{
		cfg:       config,
		year:      year,
		semester:  semester,
		schedType: 2,
		targetId:  config.studentId,
	}
}

func NewTeacherScheduleAction(config *TokenClientConfiguration, year uint, semester uint, teacher uint) *ScheduleAction {
	return &ScheduleAction{
		cfg:       config,
		year:      year,
		semester:  semester,
		schedType: 1,
		targetId:  teacher,
	}
}

func NewRoomScheduleAction(config *TokenClientConfiguration, year uint, semester uint, room uint) *ScheduleAction {
	return &ScheduleAction{
		cfg:       config,
		year:      year,
		semester:  semester,
		schedType: 4,
		targetId:  room,
	}
}

func (a *ScheduleAction) FetchSchedule() (*ics.Calendar, error) {
	req, err := a.cfg.buildRequest("POST", fmt.Sprintf(
		"/consultation/horaires/?annee=%d&trimestre=%d&type=%d&id=%d&icalendarversion=2&individual=1",
		a.year, a.semester, a.schedType, a.targetId,
	))
	if err != nil {
		return nil, err
	}

	res, err := a.cfg.doForm(req, nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	return ics.ParseCalendar(res.Body)
}
