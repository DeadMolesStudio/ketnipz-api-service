package database

import (
	"fmt"
)

type UserNotFoundError struct {
	Field string
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("no user with this %v found", e.Field)
}
