package tool

import "time"

type Service interface {
	DeaccentVietnameseString(value string) string
	GetStartOfDate(date time.Time) time.Time
	GetEndOfDate(date time.Time) time.Time
}

type service struct{}

func New() Service {
	return &service{}
}
