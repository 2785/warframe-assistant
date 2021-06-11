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

func channelMessageSender(s *discordgo.Session, cid string) MessageReplier {
	return func(msg string) error {
		_, err := s.ChannelMessageSend(cid, msg)
		return err
	}
}
