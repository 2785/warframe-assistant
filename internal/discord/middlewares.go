package discord

import (
	"fmt"

	"github.com/2785/warframe-assistant/internal/meta"
	"github.com/bwmarrin/discordgo"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
)

type MessageReplier func(msg string) error

func (h *EventHandler) mustGetOneActiveEventIDForGuild(
	gid string,
	reply MessageReplier,
) (eid string, ok bool) {
	events, err := h.MetadataService.ListActiveEventsForGuild(gid)
	if err != nil {
		h.Logger.Error("could not list events for guild", zap.Error(err), WithGuildID(gid))
		replyWithErrorLogging(
			reply,
			"Could not retrieve active events."+internalError,
			h.Logger.With(WithGuildID(gid)),
		)
		return
	}

	if len(events) == 0 {
		replyWithErrorLogging(
			reply,
			"There's no active event for this server",
			h.Logger.With(WithGuildID(gid)),
		)
		return
	} else if len(events) > 1 {
		replyWithErrorLogging(reply, "There're more than one active events for this server, please specify one of them", h.Logger.With(WithGuildID(gid)))
		return
	}

	return events[0].ID, true
}

func (h *EventHandler) mustParticipateInEvent(
	eid, uid string,
	reply MessageReplier,
) (string, bool) {
	pid, in, err := h.MetadataService.GetParticipation(uid, eid)
	if err != nil {
		if meta.AsErrNoRecord(err) {
			replyWithErrorLogging(
				reply,
				"You must be a participant of the event to perform this operation, please use `/events join` to join the event",
				h.Logger.With(WithUserID(uid), WithEventID(eid)),
			)
			return "", false
		}
		h.Logger.Error(
			"could not check if user is in event",
			zap.Error(err),
			WithUserID(uid),
			WithEventID(eid),
		)
		replyWithErrorLogging(
			reply,
			"Could not check if you signed up for the event."+internalError,
			h.Logger.With(WithUserID(uid), WithEventID(eid)),
		)
		return "", false
	}

	if !in {
		replyWithErrorLogging(
			reply,
			"You must be a participant of the event to perform this operation, please use `/events join` to join the event",
			h.Logger.With(WithUserID(uid), WithEventID(eid)),
		)
		return "", false
	}

	return pid, true
}

func (h *EventHandler) mustHaveIGNRegistered(uid string, reply MessageReplier) (string, bool) {
	ign, err := h.MetadataService.GetIGN(uid)
	if err != nil {
		if meta.AsErrNoRecord(err) {
			replyWithErrorLogging(
				reply,
				"You need to have your IGN registered with the bot to perform this action, please use `/ign register` to register",
				h.Logger.With(WithUserID(uid)),
			)
			return "", false
		}
		replyWithErrorLogging(
			reply,
			"Could not check if your IGN is registered with the bot."+internalError,
			h.Logger.With(WithUserID(uid)),
		)
		return "", false
	}

	return ign, true
}

func (h *EventHandler) mustHaveRoleWithID(
	uid, rid, gid string,
	reply MessageReplier,
	s *discordgo.Session,
) bool {
	if rid == "" {
		return true
	}
	member, err := s.GuildMember(gid, uid)
	logger := h.Logger.With(WithUserID(uid), WithGuildID(gid), WithRoleID(rid))
	if err != nil {
		logger.Error("could not get guild member information", zap.Error(err))
		replyWithErrorLogging(
			reply,
			"Could not get your guild member information."+internalError,
			logger,
		)
		return false
	}

	if !funk.Contains(member.Roles, rid) {
		roleName := "Unknown Role"

		guildRoles, err := s.GuildRoles(gid)
		if err != nil {
			logger.Error("could not retrieve roles for guild", zap.Error(err))
		} else {
			for _, r := range guildRoles {
				if r.ID == rid {
					roleName = r.Name
				}
			}

			if roleName == "Unknown Role" {
				logger.Error("role not found in guild")
			}
		}

		replyWithErrorLogging(
			reply,
			fmt.Sprintf(
				"Sorry, only users with the role '%s' in this server can perform the action",
				roleName,
			),
			logger,
		)
		return false
	}

	return true
}

func replyWithErrorLogging(reply MessageReplier, msg string, logger *zap.Logger) {
	err := reply(msg)
	if err != nil {
		logger.Error("could not reply message", zap.Error(err))
	}
}
