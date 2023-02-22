package parser

import (
	"fmt"
	"regexp"
	"strconv"
)

func (s *Parser) StudentId() (uint, error) {
	re, err := regexp.Compile(`(?m)const DEFAULT_STUDENT_ID = (\d+?);`)
	if err != nil {
		return 0, err
	}

	matches := re.FindStringSubmatch(s.src)
	if len(matches) != 2 {
		return 0, fmt.Errorf("could not find student id in javascript")
	}

	res, err := strconv.ParseUint(matches[1], 10, 32)
	return uint(res), err
}
