package team

import (
	"errors"
	"reviewer-service/internal/domain/team"
	"reviewer-service/internal/domain/user"
	"reviewer-service/internal/storage"
	storageUser "reviewer-service/internal/storage/postgresql/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func ToEntity(team *team.Model) *Entity {
	return &Entity{
		ID:   team.ID,
		Name: team.Name,
	}
}

func ToDomain(teamEntity *Entity, members []*user.Model) *team.Model {
	return &team.Model{
		ID:      teamEntity.ID,
		Name:    teamEntity.Name,
		Members: members,
	}
}

func ToDomainFromJoinResult(teamID int64, teamName string, members []*storageUser.Entity) *team.Model {
	memberModels := make([]*user.Model, 0, len(members))
	for _, member := range members {
		if member != nil && member.ID != 0 {
			memberModels = append(memberModels, storageUser.ToDomain(member))
		}
	}

	return &team.Model{
		ID:      teamID,
		Name:    teamName,
		Members: memberModels,
	}
}

func MapPGError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return storage.ErrTeamNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return storage.ErrTeamNameAlreadyExists
		case "23503":
			return storage.ErrTeamNotFound
		}
	}
	return err
}
