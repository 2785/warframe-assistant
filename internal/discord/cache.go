package discord

import (
	"errors"
	"strings"

	"github.com/2785/warframe-assistant/internal/cache"
	"github.com/bwmarrin/discordgo"
)

type dialogInfo struct {
	T                                 dialogType
	GID, MID, CID, SID, EID, CacheKey string
}

func setDialogCache(c cache.Cache, key string, info *dialogInfo) error {
	return c.Set(key, info)
}

func getDialogCache(c cache.Cache, key string) (*dialogInfo, bool) {
	info := &dialogInfo{}
	err := c.Get(key, info)

	if err != nil {
		if cache.AsErrNoRecord(err) {
			return nil, false
		}
		// potentially need to get downstream to handle it proper
		return nil, false
	}

	return info, true
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
