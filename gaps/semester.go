package gaps

import "fmt"

type Semester string

const (
	All    Semester = "all"
	First  Semester = "S1"
	Second Semester = "S2"
)

func (s Semester) String() string {
	return string(s)
}

func (s *Semester) Set(s2 string) error {
	switch s2 {
	case "":
		*s = All
		return nil
	case "all", "S1", "S2":
		*s = Semester(s2)
		return nil
	default:
		return fmt.Errorf("invalid semester: %s. Must be one of: all, S1, S2", s2)
	}
}

func (s Semester) Type() string {
	return "Semester"
}

func (s Semester) rsArg() int {
	switch s {
	case First:
		return 0
	case Second:
		return 1
	default:
		return -1
	}
}
