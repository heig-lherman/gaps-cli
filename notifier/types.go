package notifier

type ApiGrade struct {
	Course string  `json:"course"`
	Class  string  `json:"class"`
	Year   uint32  `json:"year"`
	Name   string  `json:"name"`
	Mean   float32 `json:"class_average"`
}
