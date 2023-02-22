package util

import (
	log "github.com/sirupsen/logrus"
	"reflect"
)

func CheckErr(msg any) {
	if msg != nil {
		log.Fatalf("Error: %v", msg)
	}
}

func CheckErrExcept(msg any, except any) {
	if msg != nil && reflect.TypeOf(msg) != reflect.TypeOf(except) {
		log.Fatalf("Error: %v", msg)
	}
}
