package discord

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/2785/warframe-assistant/internal/scores"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type Cache interface {
	Set(key string, val interface{}) error
	Get(key string) (interface{}, bool)
}

type MessageCreateHandler struct {
	Cache             Cache
	Logger            *zap.Logger
	Prefix            string
	EventScoreService scores.ScoresService
}

func (h *MessageCreateHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
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

func (h *MessageCreateHandler) handlePing(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Pong")
	if err != nil {
		h.Logger.Error("could not send message", zap.Error(err))
	}
}

func (h *MessageCreateHandler) handleUnknownCmd(s *discordgo.Session, m *discordgo.MessageCreate, cmd string) {
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown command '%s'", cmd))
	if err != nil {
		h.Logger.Error("could not send message", zap.Error(err))
	}
}

var submitScoreRe = regexp.MustCompile(`(?i)ign:\s*(?P<ign>\S+)\s*score:\s*(?P<score>\d+)`)

func (h *MessageCreateHandler) handleSubmitScore(s *discordgo.Session, m *discordgo.MessageCreate, text string) {
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
