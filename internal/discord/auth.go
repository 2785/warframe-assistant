package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
)

func userHasRole(s *discordgo.Session, uid, gid string, role string) (bool, error) {
	user, err := s.GuildMember(gid, uid)
	if err != nil {
		return false, err
	}

	return funk.Contains(user.Roles, role), nil
}

func (h *EventHandler) ensureUserHasRole(s *discordgo.Session, uid, gid, cid, mid, role, action string) bool {
	hasRole, err := userHasRole(s, uid, gid, role)
	if err != nil {
		h.Logger.Error("could not check if user has role", zap.Error(err))
		h.sendReplyWithLogging(s, gid, cid, mid, "Something went wrong while checking the roles of user, please try again later or contact bot maintainer for help!")
		return false
	}

	if !hasRole {
		roleName := "Unknown Role"

		guildRoles, err := s.GuildRoles(gid)

		if err != nil {
			h.Logger.Error("could not retrieve roles for guild", zap.Error(err), zap.String("gid", gid))
		} else {
			for _, r := range guildRoles {
				if r.ID == role {
					roleName = r.Name
				}
			}
			if roleName == "Unknown Role" {
				h.Logger.Error("role not found in guild", zap.String("gid", gid), zap.String("role-id", role))
			}
		}

		h.sendReplyWithLogging(s, gid, cid, mid, fmt.Sprintf("Sorry, only users with the role '%s' can perform the %s action", roleName, action))
		return false
	}

	return true
}

func (h *EventHandler) ensureUserHasRoleInteraction(s *discordgo.Session, uid, gid, cid, role, action string, i *discordgo.Interaction) bool {
	hasRole, err := userHasRole(s, uid, gid, role)
	if err != nil {
		h.Logger.Error("could not check if user has role", zap.Error(err))
		h.interactionRespondWithErrorLogging(s, i, "Something went wrong while checking the roles of user, please try again later or contact bot maintainer for help!")
		return false
	}

	if !hasRole {
		roleName := "Unknown Role"

		guildRoles, err := s.GuildRoles(gid)

		if err != nil {
			h.Logger.Error("could not retrieve roles for guild", zap.Error(err), zap.String("gid", gid))
		} else {
			for _, r := range guildRoles {
				if r.ID == role {
					roleName = r.Name
				}
			}
			if roleName == "Unknown Role" {
				h.Logger.Error("role not found in guild", zap.String("gid", gid), zap.String("role-id", role))
			}
		}

		h.interactionRespondWithErrorLogging(s, i, fmt.Sprintf("Sorry, only users with the role '%s' can perform the %s action", roleName, action))
		return false
	}

	return true
}
