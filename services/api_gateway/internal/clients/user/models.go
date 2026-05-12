package user

import (
	"encoding/json"
	"fmt"
)

type UserRole string

const (
	Admin     UserRole = "admin"
	Scientist UserRole = "scientist"
	Viewer    UserRole = "viewer"
)

func (s UserRole) IsValid() bool {
	switch s {
	case Admin, Scientist, Viewer:
		return true
	}
	return false
}

func (s *UserRole) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status := UserRole(str)
	if !status.IsValid() {
		return fmt.Errorf("invalid model status: %s", str)
	}

	*s = status
	return nil
}

type User struct {
	Email    string
	Username string
	Role     UserRole
	ID       *int
}

type RegisterCommand struct {
	Email    string
	Username string
	Password string
	Role     UserRole
}

type UserRoleUpdateCommand struct {
	Role UserRole
}

type Token struct {
	AccessToken string
	TokenType   string
}

type InternalUser struct {
	ID   int
	Role UserRole
}
