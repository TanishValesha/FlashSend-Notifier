package internal

import (
	"fmt"
	"time"
)

func ParseDateTime(input string) (*time.Time, error) {
	parsed, err := time.Parse("2006-01-02 15:04 MST", input)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime format, expected YYYY-MM-DD HH:mm, got: %s", input)
	}
	return &parsed, nil
}
