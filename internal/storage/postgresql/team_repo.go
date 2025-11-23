package postgresql

import (
	"context"
	"reviewer-service/internal/domain/team"
	storageTeam "reviewer-service/internal/storage/postgresql/team"
	storageUser "reviewer-service/internal/storage/postgresql/user"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateTeam(ctx context.Context, t *team.Model) (int64, error) {
	entity := storageTeam.ToEntity(t)

	var id int64
	var err error

	tx, pool, hasTx := s.getTx(ctx)
	if hasTx {
		err = tx.QueryRow(
			ctx,
			"INSERT INTO team (name) VALUES ($1) RETURNING id",
			entity.Name,
		).Scan(&id)
	} else {
		err = pool.QueryRow(
			ctx,
			"INSERT INTO team (name) VALUES ($1) RETURNING id",
			entity.Name,
		).Scan(&id)
	}

	if err != nil {
		return 0, storageTeam.MapPGError(err)
	}

	return id, nil
}

func (s *Storage) GetTeam(ctx context.Context, id int64) (*team.Model, error) {
	tx, pool, hasTx := s.getTx(ctx)

	query := `
		SELECT 
			team.id, 
			team.name, 
			users.id, 
			users.user_id, 
			users.username, 
			users.team_name, 
			users.is_active 
		FROM team 
		LEFT JOIN users ON team.name = users.team_name 
		WHERE team.id = $1
		ORDER BY users.id
	`

	var rows pgx.Rows
	var err error

	if hasTx {
		rows, err = tx.Query(ctx, query, id)
	} else {
		rows, err = pool.Query(ctx, query, id)
	}

	if err != nil {
		return nil, storageTeam.MapPGError(err)
	}
	defer rows.Close()

	var teamID int64
	var teamName string
	var members []*storageUser.Entity

	for rows.Next() {
		var userID *int64
		var userUserId *string
		var userUsername *string
		var userTeamName *string
		var userIsActive *bool

		err := rows.Scan(
			&teamID,
			&teamName,
			&userID,
			&userUserId,
			&userUsername,
			&userTeamName,
			&userIsActive,
		)

		if err != nil {
			return nil, storageTeam.MapPGError(err)
		}

		if userID != nil {
			members = append(members, &storageUser.Entity{
				ID:       *userID,
				UserId:   *userUserId,
				Username: *userUsername,
				TeamName: *userTeamName,
				IsActive: *userIsActive,
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, storageTeam.MapPGError(err)
	}

	if teamID == 0 {
		return nil, storageTeam.MapPGError(pgx.ErrNoRows)
	}

	return storageTeam.ToDomainFromJoinResult(teamID, teamName, members), nil
}

func (s *Storage) GetActiveReviewersByTeam(ctx context.Context, teamName string, excludeUserId string, limit int) ([]string, error) {
	tx, pool, hasTx := s.getTx(ctx)

	query := `
		SELECT user_id 
		FROM users 
		WHERE team_name = $1 
			AND is_active = true 
			AND user_id != $2
		ORDER BY user_id
		LIMIT $3
	`

	var rows pgx.Rows
	var err error

	if hasTx {
		rows, err = tx.Query(ctx, query, teamName, excludeUserId, limit)
	} else {
		rows, err = pool.Query(ctx, query, teamName, excludeUserId, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers = make([]string, 0)
	for rows.Next() {
		var reviewerId string
		if err := rows.Scan(&reviewerId); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerId)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviewers, nil
}

func (s *Storage) GetActiveReviewersByTeamExcluding(ctx context.Context, teamName string, excludeUserIds []string, limit int) ([]string, error) {
	tx, pool, hasTx := s.getTx(ctx)

	query := `
		SELECT user_id 
		FROM users 
		WHERE team_name = $1 
			AND is_active = true 
			AND user_id != ALL($2::text[])
		ORDER BY RANDOM()
		LIMIT $3
	`

	var rows pgx.Rows
	var err error

	if hasTx {
		rows, err = tx.Query(ctx, query, teamName, excludeUserIds, limit)
	} else {
		rows, err = pool.Query(ctx, query, teamName, excludeUserIds, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerId string
		if err := rows.Scan(&reviewerId); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerId)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviewers, nil
}

func (s *Storage) GetTeamByName(ctx context.Context, name string) (*team.Model, error) {
	tx, pool, hasTx := s.getTx(ctx)

	query := `
		SELECT 
			team.id, 
			team.name, 
			users.id, 
			users.user_id, 
			users.username, 
			users.team_name, 
			users.is_active 
		FROM team 
		LEFT JOIN users ON team.name = users.team_name 
		WHERE team.name = $1
		ORDER BY users.id
	`

	var rows pgx.Rows
	var err error

	if hasTx {
		rows, err = tx.Query(ctx, query, name)
	} else {
		rows, err = pool.Query(ctx, query, name)
	}

	if err != nil {
		return nil, storageTeam.MapPGError(err)
	}
	defer rows.Close()

	var teamID int64
	var teamName string
	var members []*storageUser.Entity

	for rows.Next() {
		var userID *int64
		var userUserId *string
		var userUsername *string
		var userTeamName *string
		var userIsActive *bool

		err := rows.Scan(
			&teamID,
			&teamName,
			&userID,
			&userUserId,
			&userUsername,
			&userTeamName,
			&userIsActive,
		)

		if err != nil {
			return nil, storageTeam.MapPGError(err)
		}

		if userID != nil {
			members = append(members, &storageUser.Entity{
				ID:       *userID,
				UserId:   *userUserId,
				Username: *userUsername,
				TeamName: *userTeamName,
				IsActive: *userIsActive,
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, storageTeam.MapPGError(err)
	}

	if teamID == 0 {
		return nil, storageTeam.MapPGError(pgx.ErrNoRows)
	}

	return storageTeam.ToDomainFromJoinResult(teamID, teamName, members), nil
}
