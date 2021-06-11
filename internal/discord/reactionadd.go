package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	emojiCheckMark  = "✔️"
	emojiCrossMark  = "❌"
	emojiThumbsDown = "👎"
	emojiThumbsUp   = "👍"
)

func (h *EventHandler) HandleMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// ignore bot's own reactions
	if r.UserID == s.State.User.ID {
		return
	}

	// check if this is a known dialog
	cacheKey := referenceToID(&discordgo.MessageReference{GuildID: r.GuildID, ChannelID: r.ChannelID, MessageID: r.MessageID})

	dialogInfo, ok := getDialogCache(h.Cache, cacheKey)

	if !ok {
		return
	}

	switch dialogInfo.T {
	case verificationDialog:
		role, err := h.MetadataService.GetRoleRequirementForGuild(string(verificationDialog), r.GuildID)
		if err != nil {
			h.sendReplyWithLogging(s, r.GuildID, r.ChannelID, r.MessageID, "Error retrieving role requirements for action, please try again later or contact bot maintainer for help!")
			return
		}

		if h.mustHaveRoleWithID(r.UserID, role, r.GuildID, messageReplier(s, r.GuildID, r.ChannelID, r.MessageID), s) {
			h.handleVerificationDialogReaction(s, r, dialogInfo)
		}
	default:
	}

}

func messageReplier(s *discordgo.Session, gid, cid, mid string) MessageReplier {
	return func(msg string) error {
		_, err := s.ChannelMessageSendReply(cid, msg, &discordgo.MessageReference{MessageID: mid, ChannelID: cid, GuildID: gid})
		return err
	}
}

func (h *EventHandler) handleVerificationDialogReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd, d *dialogInfo) {
	logger := h.Logger.With(WithGuildID(r.GuildID), WithUserID(r.UserID), WithSubmissionID(d.SID), zap.String("emoji", r.Emoji.Name))

	switch r.Emoji.Name {
	case emojiCheckMark:
		err := h.EventScoreService.Verify(d.SID)
		replier := messageReplier(s, d.GID, d.CID, d.MID)
		if err != nil {
			logger.Error("could not verify submission", zap.Error(err))
			replyWithErrorLogging(replier, "Could not verify submission."+internalError, logger)
			return
		}

		replyWithErrorLogging(replier, fmt.Sprintf("Successfully verified submission '%s', hit %s to see the next one", d.SID, emojiThumbsUp), logger)

		err = s.MessageReactionAdd(d.CID, d.MID, emojiThumbsUp)
		if err != nil {
			h.Logger.Error("could not add emoji reaction", zap.Error(err))
			replyWithErrorLogging(replier, "Could not add reaction for going to the next entry."+internalError, logger)
			return
		}
	case emojiThumbsUp:
		err := h.Cache.Drop(d.CacheKey)
		if err != nil {
			logger.Error("could not drop cache entry", zap.Error(err), zap.String("cache-key", d.CacheKey))
		}
		h.handleGetOneUnverifiedChannel(s, r.GuildID, r.ChannelID, d.EID)
	case emojiCrossMark:
		// this caaaaaaan potentially be problematic as the success / failure ain't captured before proceeding to the rest of
		// the flow, but I can't be fucked to sort this out rn :)
		replyWithErrorLogging(messageReplier(s, d.GID, d.CID, d.MID), fmt.Sprintf("Okay, please update the score or hit %s to remove this submission", emojiThumbsDown), logger)

		err := s.MessageReactionAdd(d.CID, d.MID, emojiThumbsDown)
		if err != nil {
			logger.Error("could not add emoji reaction", zap.Error(err))
			replyWithErrorLogging(messageReplier(s, d.GID, d.CID, d.MID), "Could not add reaction for going to the next entry."+internalError, logger)
		}
	case emojiThumbsDown:
		err := h.EventScoreService.DeleteScore(d.SID)
		if err != nil {
			h.Logger.Error("could not remove entry", zap.Error(err))
			replyWithErrorLogging(messageReplier(s, d.GID, d.CID, d.MID), "Could not removing the entry."+internalError, logger)
			return
		}

		err = h.Cache.Drop(d.CacheKey)
		if err != nil {
			h.Logger.Error("could not drop cache entry", zap.Error(err), zap.String("sid", d.SID))
		}

		replyWithErrorLogging(messageReplier(s, d.GID, d.CID, d.MID), "Successfully removed the entry.", logger)
	}

}
