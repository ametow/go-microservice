package validators

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidTimeSlice = errors.New("invalid time slice")
var ErrInvalidTime = errors.New("invalid time")

func ValidateHoursSlice(hours []string) error {
	if len(hours) == 0 {
		return ErrInvalidTimeSlice
	}
	for i := 0; i < len(hours); i++ {
		hoursStr := hours[i]
		if len(hoursStr) != 11 {
			return ErrInvalidTimeSlice
		}
		hs := strings.Split(hoursStr, "-")
		if len(hs) != 2 {
			return ErrInvalidTimeSlice
		}
		for _, hour := range hs {
			_, err := time.Parse("15:04", hour)
			if err != nil {
				return ErrInvalidTime
			}
		}
	}
	return nil
}
