package team

import "reviewer-service/internal/domain/user"

type Model struct {
	ID      int64
	Name    string
	Members []*user.Model
}
