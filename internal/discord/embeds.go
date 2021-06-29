package discord

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/thoas/go-funk"
)

const (
	verifyEmbedName     = "Please react to verify"
	eidFieldName        = "Event ID"
	eventFieldName      = "Event"
	userFieldName       = "User"
	sidFieldName        = "Submission ID"
	ignFieldName        = "IGN"
	scoresFieldName     = "Scores Claimed"
	verifiedByFieldName = "Verified By"
)

type VerificationDialog struct {
	UserDisplay string
	SID         string
	EID         string
	EventName   string
	IGN         string
	Score       string
	Verified    bool
	VerifiedBy  string
	URL         string
}

func (v *VerificationDialog) ToEmbed() *discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{
		{
			Name:  eventFieldName,
			Value: v.EventName,
		},
		{
			Name:  eidFieldName,
			Value: v.EID,
		},
		{
			Name:  userFieldName,
			Value: v.UserDisplay,
		},
		{
			Name:  sidFieldName,
			Value: v.SID,
		},
		{
			Name:  ignFieldName,
			Value: v.IGN,
		},
		{
			Name:  scoresFieldName,
			Value: v.Score,
		},
	}

	if v.Verified {
		by := "Unknown"
		if v.VerifiedBy != "" {
			by = v.VerifiedBy
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: verifiedByFieldName, Value: by})
	}

	return &discordgo.MessageEmbed{
		Image:       &discordgo.MessageEmbedImage{URL: v.URL},
		Description: verifyEmbedName,
		Fields:      fields,
	}
}

func FromEmbed(embeds []*discordgo.MessageEmbed) (*VerificationDialog, error) {
	verifEmbed := funk.Find(embeds, func(i *discordgo.MessageEmbed) bool {
		return i.Description == verifyEmbedName
	}).(*discordgo.MessageEmbed)

	if verifEmbed == nil {
		return nil, errors.New("did not embed with matching description")
	}

	url := verifEmbed.Image.URL

	fieldMap := funk.Map(
		verifEmbed.Fields,
		func(f *discordgo.MessageEmbedField) (string, string) {
			return f.Name, f.Value
		},
	).(map[string]string)

	if len(fieldMap) == 0 {
		return nil, errors.New("could not parse fields")
	}

	userDisplay, ok := fieldMap[userFieldName]
	if !ok {
		return nil, fmt.Errorf("missing field: %s", userFieldName)
	}

	eid, ok := fieldMap[eidFieldName]
	if !ok {
		return nil, fmt.Errorf("missing field: %s", eidFieldName)
	}

	eName, ok := fieldMap[eventFieldName]
	if !ok {
		return nil, fmt.Errorf("missing field: %s", eventFieldName)
	}

	sid, ok := fieldMap[sidFieldName]
	if !ok {
		return nil, fmt.Errorf("missing field: %s", sidFieldName)
	}

	ign, ok := fieldMap[ignFieldName]
	if !ok {
		return nil, fmt.Errorf("missing field: %s", ignFieldName)
	}

	score, ok := fieldMap[scoresFieldName]
	if !ok {
		return nil, fmt.Errorf("missing field: %s", scoresFieldName)
	}

	out := &VerificationDialog{
		UserDisplay: userDisplay,
		SID:         sid,
		IGN:         ign,
		Score:       score,
		URL:         url,
		EID:         eid,
		EventName:   eName,
	}

	if verifiedBy, ok := fieldMap[verifiedByFieldName]; ok {
		out.Verified = true
		out.VerifiedBy = verifiedBy
	}

	return out, nil
}
