package discord

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Cache interface {
	Set(key string, val string) error
	Get(key string) (string, bool)
}

func referenceToID(r *discordgo.MessageReference) string {
	return r.GuildID + "|" + r.ChannelID + "|" + r.MessageID
}

func idToReference(s string) (*discordgo.MessageReference, error) {
	splits := strings.Split(s, "|")
	if len(splits) != 3 {
		return nil, errors.New("expected 3 segments")
	}
	return &discordgo.MessageReference{
		GuildID:   splits[0],
		ChannelID: splits[1],
		MessageID: splits[2],
	}, nil
}
