package discord

import (
	"github.com/2785/warframe-assistant/internal/cache"
	"github.com/2785/warframe-assistant/internal/meta"
	"github.com/2785/warframe-assistant/internal/scores"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type EventHandler struct {
	Cache             cache.Cache
	Logger            *zap.Logger
	Prefix            string
	EventScoreService scores.ScoresService
	MetadataService   meta.Service
}

type dialogType string

const (
	verificationDialog dialogType = dialogType("verification")
	submissionDialog   dialogType = dialogType("submission")
)

func (h *EventHandler) sendReplyWithLogging(s *discordgo.Session, gid, cid, mid, msg string) {
	_, err := s.ChannelMessageSendReply(cid, msg, &discordgo.MessageReference{MessageID: mid, ChannelID: cid, GuildID: gid})
	if err != nil {
		h.Logger.Error("could not send message", zap.Error(err))
	}
}
