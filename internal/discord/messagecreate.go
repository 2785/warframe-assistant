package discord

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/2785/warframe-assistant/internal/scores"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	emojiYes = "✔️"
	emojiNo  = "❌"
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
	case "verifyone":
		h.handleGetOne(s, m)
	case "scores":
		h.handleScoreBoard(s, m)
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

var submitScoreRe = regexp.MustCompile(`(?i)ign:\s*(?P<ign>\S+)\s*score:\s*(?P<score>\d+)`)

func (h *EventHandler) handleSubmitScore(s *discordgo.Session, m *discordgo.MessageCreate, text string) {
	userID := m.Author.Username + "#" + m.Author.Discriminator
	match := submitScoreRe.FindStringSubmatch(text)

	fixInputMsg := fmt.Sprintf("Could not understand the input, please make your submission in the format `%ssubmit ign: <your-ign> score: <your score>`, without the angle brackets", h.Prefix)

	if match == nil {
		_, err := s.ChannelMessageSendReply(m.ChannelID, fixInputMsg, m.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	inputMap := make(map[string]string)
	for i, name := range submitScoreRe.SubexpNames() {
		if i > 0 && i <= len(match) {
			inputMap[name] = match[i]
		}
	}

	for _, v := range inputMap {
		if v == "" {
			_, err := s.ChannelMessageSendReply(m.ChannelID, fixInputMsg, m.Reference())
			if err != nil {
				h.Logger.Error("could not send message", zap.Error(err))
			}
			return
		}
	}

	ign := strings.TrimSpace(inputMap["ign"])
	scoreStr := strings.TrimSpace(inputMap["score"])

	scoreInt, err := strconv.Atoi(scoreStr)
	if err != nil {
		_, err = s.ChannelMessageSendReply(m.ChannelID, "score needs to be an integer", m.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	if len(m.Attachments) != 1 {
		_, err = s.ChannelMessageSendReply(m.ChannelID, "Please provide a screenshot as evidence for the score", m.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	proof := m.Attachments[0].URL

	id, err := h.EventScoreService.AddUnverified(userID, ign, scoreInt, proof)
	if err != nil {
		h.Logger.Error("could not upload score", zap.Error(err))
		_, err = s.ChannelMessageSendReply(m.ChannelID, "Error uploading score, please try again later or contact 2785, err: "+err.Error(), m.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	_, err = s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("Successfully uploaded score for ign `%s` - %v score, submission ID is %s", ign, scoreInt, id), m.Reference())
	if err != nil {
		h.Logger.Error("could not send message", zap.Error(err))
	}
}

func (h *EventHandler) handleGetOne(s *discordgo.Session, m *discordgo.MessageCreate) {
	id, userid, ign, score, proof, err := h.EventScoreService.GetOneUnverified()
	if err != nil {
		noRecord := &scores.ErrNoRecord{}
		if errors.As(err, &noRecord) {
			_, err = s.ChannelMessageSendReply(m.ChannelID, ":tada: There are no pending submissions to be verified", m.Reference())
			if err != nil {
				h.Logger.Error("could not send message", zap.Error(err))
			}
			return
		}

		_, err = s.ChannelMessageSendReply(m.ChannelID, "Sorry, something's borked, please try again or contact bot maintainer for help!", m.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	sent, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Image:       &discordgo.MessageEmbedImage{URL: proof},
		Description: "Please react to verify",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: userid},
			{Name: "Sumbission ID", Value: id},
			{Name: "IGN", Value: "`" + ign + "`"},
			{Name: "Scores Claimed", Value: fmt.Sprintf("%v", score)},
		},
	})

	if err != nil {
		h.Logger.Error("could not send message", zap.Error(err))
		return
	}

	verificationErrMsg := "Caching / Reaction seem to be borked, reaction verification workflow would not function, please try again later or contact bot maintainer for help!"

	msgId := referenceToID(&discordgo.MessageReference{GuildID: m.GuildID, ChannelID: m.ChannelID, MessageID: sent.Reference().MessageID})
	h.Logger.Debug("storing into cache", zap.Any("cache, before", h.Cache))
	err = h.Cache.Set(msgId, id)
	if err != nil {
		h.Logger.Error("could not add entry to cache", zap.Error(err))
		_, err = s.ChannelMessageSendReply(m.ChannelID, verificationErrMsg, sent.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	h.Logger.Debug("stored into cache", zap.Any("cache, after", h.Cache))

	err = s.MessageReactionAdd(sent.Reference().ChannelID, sent.Reference().MessageID, "✔️")

	if err != nil {
		h.Logger.Error("could not add reaction", zap.Error(err))
		_, err = s.ChannelMessageSendReply(m.ChannelID, verificationErrMsg, sent.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	err = s.MessageReactionAdd(sent.Reference().ChannelID, sent.Reference().MessageID, "❌")

	if err != nil {
		h.Logger.Error("could not add reaction", zap.Error(err))
		_, err = s.ChannelMessageSendReply(m.ChannelID, verificationErrMsg, sent.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}
}

func (h *EventHandler) handleScoreBoard(s *discordgo.Session, m *discordgo.MessageCreate) {
	scoreboard, err := h.EventScoreService.ScoreReport()
	if err != nil {
		h.Logger.Error("could not fetch scoreboard", zap.Error(err))
		_, err = s.ChannelMessageSendReply(m.ChannelID, "Error fetching score board, please try again later or contact bot maintainer for help!", m.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	if len(scoreboard) == 0 {
		_, err = s.ChannelMessageSendReply(m.ChannelID, "There's no verified submission to the event yet!", m.Reference())
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}

	fields := make([]string, len(scoreboard))

	for i, v := range scoreboard {
		fields[i] = fmt.Sprintf("#%v - %s (`%s`) - %v points", i+1, v.UserID, v.IGN, v.Sum)
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Title:       "Event Scores",
		Description: strings.Join(fields, "\n"),
	})

	if err != nil {
		if err != nil {
			h.Logger.Error("could not send message", zap.Error(err))
		}
		return
	}
}
