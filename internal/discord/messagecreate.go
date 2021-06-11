package discord

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/2785/warframe-assistant/internal/scores"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func (h *EventHandler) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Author.Bot {
		return
	}

	if !strings.HasPrefix(m.Content, h.Prefix) {
		return
	}

	msg := strings.TrimSpace(strings.TrimPrefix(m.Content, h.Prefix))

	cmd := strings.Split(msg, " ")[0]

	msg = strings.TrimPrefix(msg, cmd)

	switch strings.ToLower(cmd) {
	case "ping":
		h.handlePing(s, m)
	case "submit":
		h.handleSubmitScore(s, m, msg)
	default:
		h.handleUnknownCmd(s, m, cmd)
	}
}

func (h *EventHandler) handlePing(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Pong")
	if err != nil {
		h.Logger.Error("could not send message", zap.Error(err))
	}
}

func (h *EventHandler) handleUnknownCmd(s *discordgo.Session, m *discordgo.MessageCreate, cmd string) {
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown command '%s'", cmd))
	if err != nil {
		h.Logger.Error("could not send message", zap.Error(err))
	}
}

var submitScoreRe = regexp.MustCompile(`(?i)\s*(?P<score>\d+)(\s*event:\s*(?P<event>\S+))?`)

func (h *EventHandler) handleSubmitScore(s *discordgo.Session, m *discordgo.MessageCreate, text string) {
	logger := h.Logger.With(WithGuildID(m.GuildID), WithMessageID(m.ID), WithChannelID(m.ChannelID), WithUserID(m.Author.ID), WithHandler("handle-submit-score"))
	replier := messageReplier(s, m.GuildID, m.ChannelID, m.ID)
	fixInputMsg := fmt.Sprintf("Could not understand the input, please make your submission in the format `%ssubmit ign: <your-ign> score: <your score>`, without the angle brackets", h.Prefix)

	// first we'll see if we can get the event ID
	match := submitScoreRe.FindStringSubmatch(text)

	if match == nil {
		replyWithErrorLogging(replier, fixInputMsg, logger)
		return
	}

	inputMap := make(map[string]string)
	for i, name := range submitScoreRe.SubexpNames() {
		if i > 0 && i <= len(match) {
			inputMap[name] = match[i]
		}
	}

	eid := strings.ToLower(strings.TrimSpace(inputMap["event"]))
	if eid == "" {
		var ok bool
		eid, ok = h.mustGetOneActiveEventIDForGuild(m.GuildID, replier)
		if !ok {
			return
		}
	}

	event, err := h.MetadataService.GetEvent(eid)
	if err != nil {
		replyWithErrorLogging(replier, "Error fetching event information."+internalError, logger)
		return
	}

	if event.Begin.After(time.Now()) {
		replyWithErrorLogging(replier, "Event is not open yet", logger)
		return
	}

	if event.End.Before(time.Now()) {
		replyWithErrorLogging(replier, "Event submission is already closed", logger)
		return
	}

	// now we make sure the user is in the event
	pid, ok := h.mustParticipateInEvent(eid, m.Author.ID, replier)
	if !ok {
		return
	}

	scoreStr := strings.TrimSpace(inputMap["score"])

	score, err := strconv.Atoi(scoreStr)
	if err != nil {
		replyWithErrorLogging(replier, "score needs to be an integer", logger)
		return
	}

	if len(m.Attachments) != 1 {
		replyWithErrorLogging(replier, "Please provide a screenshot as evidence for the score", logger)
		return
	}

	proof := m.Attachments[0].URL

	sid, err := h.EventScoreService.ClaimScore(pid, score, proof)

	if err != nil {
		logger.Error("could not upload score", zap.Error(err))
		replyWithErrorLogging(replier, "Error uploading score."+internalError, logger)
		return
	}

	replyWithErrorLogging(replier, fmt.Sprintf("Successfully uploaded score (%v) - submission ID is %s", score, sid), logger)
}

func (h *EventHandler) handleGetOneUnverifiedChannel(s *discordgo.Session, gid, cid, eid string) {
	record, err := h.EventScoreService.GetOneUnverifiedForEvent(eid)
	logger := h.Logger.With(WithGuildID(gid), WithHandler("handle-get-one-unverified"), WithChannelID(cid), WithEventID(eid))

	if err != nil {
		if scores.AsErrNoRecord(err) {
			replyWithErrorLogging(channelMessageSender(s, cid), ":tada: There are no pending submissions to be verified", logger)
			return
		}
		replyWithErrorLogging(channelMessageSender(s, cid), "Sorry, something's borked."+internalError, logger)
		return
	}

	sent, err := s.ChannelMessageSendEmbed(cid, &discordgo.MessageEmbed{
		Image:       &discordgo.MessageEmbedImage{URL: record.Proof},
		Description: "Please react to verify",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: record.UID},
			{Name: "Sumbission ID", Value: record.ID},
			{Name: "IGN", Value: "`" + record.IGN + "`"},
			{Name: "Scores Claimed", Value: fmt.Sprintf("%v", record.Score)},
		},
	})

	if err != nil {
		logger.Error("could not send message", zap.Error(err))
		return
	}

	verificationErrMsg := "Caching / Reaction seem to be borked, reaction verification workflow would not function." + internalError

	cacheKey := referenceToID(&discordgo.MessageReference{GuildID: gid, ChannelID: cid, MessageID: sent.Reference().MessageID})
	logger.Debug("storing into cache", zap.Any("cache, before", h.Cache))

	err = setDialogCache(h.Cache, cacheKey, &dialogInfo{
		T:        verificationDialog,
		GID:      gid,
		CID:      cid,
		MID:      sent.Reference().MessageID,
		SID:      record.ID,
		EID:      eid,
		CacheKey: cacheKey,
	})

	if err != nil {
		logger.Error("could not add entry to cache", zap.Error(err))
		replyWithErrorLogging(messageReplier(s, gid, cid, sent.Reference().MessageID), verificationErrMsg, logger)
		return
	}

	err = s.MessageReactionAdd(sent.Reference().ChannelID, sent.Reference().MessageID, "✔️")

	if err != nil {
		h.Logger.Error("could not add reaction", zap.Error(err))
		replyWithErrorLogging(messageReplier(s, gid, cid, sent.Reference().MessageID), verificationErrMsg, logger)
		return
	}

	err = s.MessageReactionAdd(sent.Reference().ChannelID, sent.Reference().MessageID, "❌")

	if err != nil {
		h.Logger.Error("could not add reaction", zap.Error(err))
		replyWithErrorLogging(messageReplier(s, gid, cid, sent.Reference().MessageID), verificationErrMsg, logger)
		return
	}
}

// func (h *EventHandler) handleScoreBoard(s *discordgo.Session, m *discordgo.MessageCreate) {
// 	scoreboard, err := h.EventScoreService.ScoreReport()
// 	if err != nil {
// 		h.Logger.Error("could not fetch scoreboard", zap.Error(err))
// 		_, err = s.ChannelMessageSendReply(m.ChannelID, "Error fetching score board, please try again later or contact bot maintainer for help!", m.Reference())
// 		if err != nil {
// 			h.Logger.Error("could not send message", zap.Error(err))
// 		}
// 		return
// 	}

// 	if len(scoreboard) == 0 {
// 		_, err = s.ChannelMessageSendReply(m.ChannelID, "There's no verified submission to the event yet!", m.Reference())
// 		if err != nil {
// 			h.Logger.Error("could not send message", zap.Error(err))
// 		}
// 		return
// 	}

// 	fields := make([]string, len(scoreboard))

// 	for i, v := range scoreboard {
// 		fields[i] = fmt.Sprintf("#%v - %s (`%s`) - %v points", i+1, v.UserID, v.IGN, v.Sum)
// 	}

// 	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
// 		Title:       "Event Scores",
// 		Description: strings.Join(fields, "\n"),
// 	})

// 	if err != nil {
// 		if err != nil {
// 			h.Logger.Error("could not send message", zap.Error(err))
// 		}
// 		return
// 	}
// }

// func (h *EventHandler) handleCheckVerificationProgress(s *discordgo.Session, m *discordgo.MessageCreate) {
// 	t, v, p, err := h.EventScoreService.VerificationStatus()
// 	if err != nil {
// 		h.sendReplyWithLogging(s, m.GuildID, m.ChannelID, m.ID, "Could not get verification status, please try again later or contact bot maintainer for help!")
// 		return
// 	}

// 	h.sendReplyWithLogging(s, m.GuildID, m.ChannelID, m.ID, fmt.Sprintf("%v/%v done, %s", v, t, func() string {
// 		if p != 0 {
// 			return fmt.Sprintf("%v to go.", p)
// 		}
// 		return "all done! :tada:"
// 	}()))
// }

// var updateScoreRe = regexp.MustCompile(`(?i)id:\s*(?P<id>\S+)\s*score:\s*(?P<score>\d+)`)

// func (h *EventHandler) handleUpdateScore(s *discordgo.Session, m *discordgo.MessageCreate, text string) {
// 	match := updateScoreRe.FindStringSubmatch(text)

// 	fixInputMsg := fmt.Sprintf("Could not understand the input, please update score in the format `%supdate id: <id of the submission> score: <new score>`, without the angle brackets", h.Prefix)

// 	if match == nil {
// 		h.sendReplyWithLogging(s, m.GuildID, m.ChannelID, m.ID, fixInputMsg)
// 		return
// 	}

// 	inputMap := make(map[string]string)
// 	for i, name := range updateScoreRe.SubexpNames() {
// 		if i > 0 && i <= len(match) {
// 			inputMap[name] = match[i]
// 		}
// 	}

// 	for _, v := range inputMap {
// 		if v == "" {
// 			h.sendReplyWithLogging(s, m.GuildID, m.ChannelID, m.ID, fixInputMsg)
// 			return
// 		}
// 	}

// 	id := strings.TrimSpace(inputMap["id"])
// 	score := strings.TrimSpace(inputMap["score"])

// 	scoreInt, err := strconv.Atoi(score)
// 	if err != nil {
// 		h.sendReplyWithLogging(s, m.GuildID, m.ChannelID, m.ID, "Scores has to be a valid integer!")
// 		return
// 	}

// 	err = h.EventScoreService.UpdateScore(id, scoreInt)
// 	if err != nil {
// 		h.Logger.Error("could not update score", zap.Error(err), zap.String("sid", id))
// 		h.sendReplyWithLogging(s, m.GuildID, m.ChannelID, m.ID, "Could not update score, please try again later or contact bot maintainer for help!")
// 		return
// 	}
// 	h.sendReplyWithLogging(s, m.GuildID, m.ChannelID, m.ID, "Successfully updated score and updated status to verified!")

// }

func channelMessageSender(s *discordgo.Session, cid string) MessageReplier {
	return func(msg string) error {
		_, err := s.ChannelMessageSend(cid, msg)
		return err
	}
}
