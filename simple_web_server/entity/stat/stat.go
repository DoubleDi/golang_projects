package stat

import (
	"fmt"
	"time"
)

type timeWithoutSeconds time.Time

func (t *timeWithoutSeconds) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	time, err := time.Parse(`"`+DatetimeLayout+`"`, string(data))
	*t = timeWithoutSeconds(time)

	return err
}

type Stat struct {
	UserID int                `json:"user"`
	Action string             `json:"action"`
	TS     timeWithoutSeconds `json:"ts"`
}

var AvailibleActions = map[string]struct{}{
	"like":    struct{}{},
	"comment": struct{}{},
	"exit":    struct{}{},
	"login":   struct{}{},
}

const (
	DatetimeLayout = "2006-01-02T15:04:05"
	DateLayout     = "2006-01-02"
)

func (s *Stat) Validate() error {
	if !IsValidAction(s.Action) {
		return fmt.Errorf("not availible action %v", s.Action)
	}

	return nil
}

func IsValidAction(action string) bool {
	_, exists := AvailibleActions[action]

	return exists
}
