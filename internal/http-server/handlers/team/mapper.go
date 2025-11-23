package team

import (
	"reviewer-service/internal/domain/team"
	"reviewer-service/internal/domain/user"
)

func toDomain(dto *DTO) *team.Model {
	return &team.Model{
		Name:    dto.Name,
		Members: toUserDomains(dto.Members, dto.Name),
	}
}

func toDto(teamModel *team.Model) *DTO {
	return &DTO{
		Name:    teamModel.Name,
		Members: toMemberDTOs(teamModel.Members),
	}
}

func toUserDomains(dtos []*Member, teamName string) []*user.Model {
	users := make([]*user.Model, len(dtos))
	for i, dto := range dtos {
		users[i] = &user.Model{
			UserId:   dto.UserId,
			Username: dto.Username,
			TeamName: teamName,
			IsActive: dto.IsActive,
		}
	}
	return users
}

func toMemberDTOs(users []*user.Model) []*Member {
	members := make([]*Member, len(users))
	for i, userModel := range users {
		members[i] = &Member{
			UserId:   userModel.UserId,
			Username: userModel.Username,
			IsActive: userModel.IsActive,
		}
	}
	return members
}
