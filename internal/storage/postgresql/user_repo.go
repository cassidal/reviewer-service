package postgresql

import (
	"context"
	"reviewer-service/internal/domain/user"
	storageUser "reviewer-service/internal/storage/postgresql/user"
)

func (s *Storage) CreateUser(ctx context.Context, u *user.Model) (int64, error) {
	entity := storageUser.ToEntity(u)

	var id int64
	var err error

	tx, pool, hasTx := s.getTx(ctx)

	sql := `
		INSERT INTO users 
    		(user_id, username, team_name, is_active) 
		VALUES 
    		($1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
		RETURNING id
	`
	if hasTx {
		err = tx.QueryRow(
			ctx,
			sql,
			entity.UserId,
			entity.Username,
			entity.TeamName,
			entity.IsActive,
		).Scan(&id)
	} else {
		err = pool.QueryRow(
			ctx,
			sql,
			entity.UserId,
			entity.Username,
			entity.TeamName,
			entity.IsActive,
		).Scan(&id)
	}

	if err != nil {
		return 0, storageUser.MapPGError(err)
	}

	return id, nil
}

func (s *Storage) GetUser(ctx context.Context, id int64) (*user.Model, error) {
	var entity storageUser.Entity
	var err error

	tx, pool, hasTx := s.getTx(ctx)
	sql := "SELECT id, user_id, username, team_name, is_active FROM users WHERE id=$1"
	if hasTx {
		err = tx.QueryRow(
			ctx,
			sql,
			id,
		).Scan(&entity.ID, &entity.UserId, &entity.Username, &entity.TeamName, &entity.IsActive)
	} else {
		err = pool.QueryRow(
			ctx,
			sql,
			id,
		).Scan(&entity.ID, &entity.UserId, &entity.Username, &entity.TeamName, &entity.IsActive)
	}

	if err != nil {
		return nil, storageUser.MapPGError(err)
	}

	return storageUser.ToDomain(&entity), nil
}

func (s *Storage) GetUserByUserId(ctx context.Context, userId string) (*user.Model, error) {
	var entity storageUser.Entity
	var err error

	tx, pool, hasTx := s.getTx(ctx)
	sql := "SELECT id, user_id, username, team_name, is_active FROM users WHERE user_id=$1"
	if hasTx {
		err = tx.QueryRow(
			ctx,
			sql,
			userId,
		).Scan(&entity.ID, &entity.UserId, &entity.Username, &entity.TeamName, &entity.IsActive)
	} else {
		err = pool.QueryRow(
			ctx,
			sql,
			userId,
		).Scan(&entity.ID, &entity.UserId, &entity.Username, &entity.TeamName, &entity.IsActive)
	}

	if err != nil {
		return nil, storageUser.MapPGError(err)
	}

	return storageUser.ToDomain(&entity), nil
}

func (s *Storage) UpdateUserIsActive(ctx context.Context, userId string, isActive bool) (int64, error) {
	tx, pool, hasTx := s.getTx(ctx)

	sql := `
		UPDATE users 
		SET is_active = $1 
		WHERE user_id = $2 
		RETURNING id
	`

	var id int64
	var err error

	if hasTx {
		err = tx.QueryRow(
			ctx,
			sql,
			isActive,
			userId,
		).Scan(&id)
	} else {
		err = pool.QueryRow(
			ctx,
			sql,
			isActive,
			userId,
		).Scan(&id)
	}

	if err != nil {
		return 0, storageUser.MapPGError(err)
	}

	return id, nil
}
