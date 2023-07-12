package validators

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateHoursSlice(t *testing.T) {
	tcases := []struct {
		name   string
		input  []string
		expect error
	}{
		{
			name:   "empty input",
			input:  []string{},
			expect: ErrInvalidTimeSlice,
		},
		{
			name:   "invalid times",
			input:  []string{"ab", "bc"},
			expect: ErrInvalidTimeSlice,
		},
		{
			name:   "omit leading zero",
			input:  []string{"1:00-13:09", "00:14-14:00"},
			expect: ErrInvalidTimeSlice,
		},
		{
			name:   "invalid delimeter in timeslice",
			input:  []string{"01:00_13:09", "00:14-14:00"},
			expect: ErrInvalidTimeSlice,
		},
		{
			name:   "mistaken alphabet",
			input:  []string{"o1:00-13:09", "00:14-14:00"},
			expect: ErrInvalidTime,
		},
		{
			name:   "success case",
			input:  []string{"01:00-13:09", "00:14-14:00"},
			expect: nil,
		},
	}
	for _, tcase := range tcases {
		t.Run(tcase.name, func(t *testing.T) {
			err := ValidateHoursSlice(tcase.input)
			require.Equal(t, tcase.expect, err)
		})
	}
}
