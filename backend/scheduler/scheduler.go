package scheduler

import (
	"log"
	"time"

	"github.com/rjxby/eat-repeat/backend/store"
)

const DayFormat = "2006-01-02"
const DayTitleFormat = "January 02, Monday"

// ScheduleProc creates and save schedules
type ScheduleProc struct {
}

// New makes ScheduleProc
func New() *ScheduleProc {
	return &ScheduleProc{}
}

func generateWeek(offsetInDays int) (week *store.Week, err error) {
	currentTime := time.Now()

	startOfWeek := currentTime.AddDate(0, 0, -int(currentTime.Weekday())+offsetInDays)

	var days []store.Day

	for i := 0; i < 7; i++ {
		day := startOfWeek.AddDate(0, 0, i)

		dayToAdd := store.Day{
			ID:           day.Format(DayFormat),
			Title:        day.Format(DayTitleFormat),
			IsCurrentDay: day.Format(DayFormat) == currentTime.Format(DayFormat),
		}

		days = append(days, dayToAdd)
	}

	year, number := startOfWeek.ISOWeek()

	week = &store.Week{
		Days:   days,
		Number: number,
		Year:   year,
	}

	log.Printf("[INFO] week is generated: %v", week)

	return week, nil
}

func (p ScheduleProc) GetWeek() (week *store.Week, err error) {
	return generateWeek(0)
}

func (p ScheduleProc) GetNextWeek() (week *store.Week, err error) {
	return generateWeek(7)
}
