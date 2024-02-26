package notifier

type ApiGrade struct {
	Course string  `json:"course"`
	Class  string  `json:"class"`
	Name   string  `json:"name"`
	Mean   float32 `json:"class_average"`
}
