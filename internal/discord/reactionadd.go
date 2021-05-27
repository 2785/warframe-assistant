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
	msgId := referenceToID(&discordgo.MessageReference{GuildID: r.GuildID, ChannelID: r.ChannelID, MessageID: r.MessageID})

	dialogInfo, ok := getDialogCache(h.Cache, msgId)

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

		if h.ensureUserHasRole(s, r.UserID, r.GuildID, r.ChannelID, r.MessageID, role, "submission verification") {
			h.handleVerificationDialogReaction(s, r, dialogInfo)
		}
	default:
	}

}

func (h *EventHandler) handleVerificationDialogReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd, d *dialogInfo) {
	switch r.Emoji.Name {
	case emojiCheckMark:
		err := h.EventScoreService.Verify(d.SID)
		if err != nil {
			h.Logger.Error("could not verify submission", zap.Error(err), zap.String("sid", d.SID), zap.String("cid", d.CID), zap.String("mid", d.MID))
			h.sendReplyWithLogging(s, d.GID, d.CID, d.MID, "Could not verify submission, please try again later or contact bot maintainer for help!")
			return
		}

		h.sendReplyWithLogging(s, d.GID, d.CID, d.MID, fmt.Sprintf("Successfully verified submission '%s', hit %s to see the next one", d.SID, emojiThumbsUp))

		err = s.MessageReactionAdd(d.CID, d.MID, emojiThumbsUp)
		if err != nil {
			h.Logger.Error("could not add emoji reaction", zap.Error(err))
			h.sendReplyWithLogging(s, d.GID, d.CID, d.MID, "Could not add reaction for going to the next entry, please try again later or contact bot maintainer for help")
			return
		}
	case emojiThumbsUp:
		err := h.Cache.Drop(d.SID)
		if err != nil {
			h.Logger.Error("could not drop cache entry", zap.Error(err), zap.String("sid", d.SID))
		}
		h.handleGetOne(s, r.GuildID, r.ChannelID, "")
	case emojiCrossMark:
		h.sendReplyWithLogging(s, d.GID, d.CID, d.MID, fmt.Sprintf("Okay, please update the score or hit %s to remove this submission", emojiThumbsDown))

		err := s.MessageReactionAdd(d.CID, d.MID, emojiThumbsDown)
		if err != nil {
			h.Logger.Error("could not add emoji reaction", zap.Error(err))
			h.sendReplyWithLogging(s, d.GID, d.CID, d.MID, "Could not add reaction for going to the next entry, please try again later or contact bot maintainer for help")
			return
		}
	case emojiThumbsDown:

		err := h.EventScoreService.DeleteRecord(d.SID)
		if err != nil {
			h.Logger.Error("could not remove entry", zap.Error(err))
			h.sendReplyWithLogging(s, d.GID, d.CID, d.MID, "error removing the entry, please try again later or contact bot maintainer for help!")
			return
		}

		err = h.Cache.Drop(d.SID)
		if err != nil {
			h.Logger.Error("could not drop cache entry", zap.Error(err), zap.String("sid", d.SID))
		}

		h.sendReplyWithLogging(s, d.GID, d.CID, d.MID, "Successfully deleted the submission")
	}

}
