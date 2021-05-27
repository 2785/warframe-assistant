package discord

import (
	"fmt"

	"github.com/2785/warframe-assistant/internal/scores"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type EventHandler struct {
	Cache                 Cache
	Logger                *zap.Logger
	Prefix                string
	EventScoreService     scores.ScoresService
	RoleIDForVerification string
}

func (h *EventHandler) HandleMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	h.Logger.Info("caught a thing", zap.Any("r", r), zap.Any("cache", h.Cache))

	if r.UserID == s.State.User.ID {
		return
	}

	msgId := referenceToID(&discordgo.MessageReference{GuildID: r.GuildID, ChannelID: r.ChannelID, MessageID: r.MessageID})

	submissionID, ok := h.Cache.Get(msgId)

	if !ok {
		return
	}

	h.Logger.Debug("got a message reaction thingy", zap.Any("data", r))

	_ = submissionID

	user, err := s.GuildMember(r.GuildID, r.UserID)
	if err != nil {
		_, err = s.ChannelMessageSendReply(r.ChannelID, "Could not fetch user information, please try again later or contact bot maintainer for help!", &discordgo.MessageReference{GuildID: r.GuildID, ChannelID: r.ChannelID, MessageID: r.MessageID})
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	h.Logger.Debug("check user roles", zap.String("roles", fmt.Sprintf("%+v", user.Roles)))

	if !func() bool {
		for _, v := range user.Roles {
			if v == h.RoleIDForVerification {
				return true
			}
		}
		return false
	}() {
		_, err = s.ChannelMessageSendReply(r.ChannelID, fmt.Sprintf("Sorry %s, only members with the %s role can verify submissions", user.Mention(), h.RoleIDForVerification), &discordgo.MessageReference{GuildID: r.GuildID, ChannelID: r.ChannelID, MessageID: r.MessageID})

		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	if r.Emoji.Name == emojiYes {
		err = h.EventScoreService.Verify(submissionID)
		if err != nil {
			_, err = s.ChannelMessageSendReply(r.ChannelID, fmt.Sprintf("Sorry %s, there was an error while verifying the record, please try again in a bit or contact bot maintainer for help!", user.Mention()), &discordgo.MessageReference{GuildID: r.GuildID, ChannelID: r.ChannelID, MessageID: r.MessageID})
			if err != nil {
				h.Logger.Error("could not send message", zap.Error(err))
			}
			return
		}
		_, err = s.ChannelMessageSendReply(r.ChannelID, fmt.Sprintf("Successfully verified submission with ID %s by %s", submissionID, user.Mention()), &discordgo.MessageReference{GuildID: r.GuildID, ChannelID: r.ChannelID, MessageID: r.MessageID})
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
	}
}
