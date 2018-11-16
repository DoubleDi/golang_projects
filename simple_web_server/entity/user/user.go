package user

import "fmt"

type User struct {
	ID    int    `json:"id"`
	Age   int    `json:"age"`
	Sex   string `json:"sex"`
	Count int    `json:"count"`
}

type UsersByDate struct {
	Date string `json:"date"`
	Rows []User `json:"rows"`
}

var availibleSex = map[string]struct{}{
	"M": struct{}{},
	"F": struct{}{},
}

func (u *User) Validate() error {
	if _, exists := availibleSex[u.Sex]; !exists {
		return fmt.Errorf("not availible sex %v", u.Sex)
	}

	return nil
}
