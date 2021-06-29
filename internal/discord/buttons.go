package discord

import (
	"fmt"

	"github.com/2785/warframe-assistant/internal/scores"
	"github.com/bwmarrin/discordgo"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
)

const (
	scoreVerificationBotton = "score-verify-btn"
	scoreRejectButton       = "score-reject-btn"
	scoreNextButton         = "score-next-btn"
	scoreRemoveButton       = "score-remove-btn"
)

var buttonAuth map[string]string = map[string]string{
	scoreVerificationBotton: string(verificationDialog),
	scoreRejectButton:       string(verificationDialog),
	scoreNextButton:         string(verificationDialog),
	scoreRemoveButton:       string(verificationDialog),
}

func (h *EventHandler) handleInteractionButtons(
	s *discordgo.Session,
	i *discordgo.Interaction,
	btn string,
) {
	logger := h.Logger.With(
		WithComponent("interaction-button-handler"),
		WithGuildID(i.GuildID),
		zap.String("button-id", btn),
		WithUserID(i.Member.User.ID),
		WithChannelID(i.ChannelID),
	)

	logger.Debug("handleInteractionButtons", zap.String("btn", btn))

	// handle auth - now admittedly this auth should probably be taken from a config file as opposed
	// to hard coded in code.
	if r, ok := buttonAuth[btn]; ok {
		rid, err := h.MetadataService.GetRoleRequirementForGuild(r, i.GuildID)
		if err != nil {
			replyWithErrorLogging(
				interactionReplier(s, i),
				"Could not retrieve role requirement for action."+internalError,
				logger.With(zap.String("dialog-id", r)),
			)
			return
		}

		if !h.mustHaveRoleWithID(i.Member.User.ID, rid, i.GuildID, interactionReplier(s, i), s) {
			return
		}
	}

	// extract submission information for each of the verification dialog button options
	if funk.Contains(
		[]string{scoreVerificationBotton, scoreRejectButton, scoreNextButton, scoreRemoveButton},
		btn,
	) {
		dialog, err := FromEmbed(i.Message.Embeds)
		if err != nil {
			logger.Error("could not parse message embeds", zap.Error(err))
			replyWithErrorLogging(
				interactionReplier(s, i),
				"Error parsing message."+internalError,
				logger,
			)
			return
		}

		switch btn {
		case scoreVerificationBotton:
			h.handleVerifyButton(dialog, s, i, logger)
		case scoreRejectButton:
			h.handleRejectButton(dialog, s, i, logger)
		case scoreNextButton:
			h.handleNextButton(dialog, s, i, logger)
		case scoreRemoveButton:
			h.handleRemoveButton(dialog, s, i, logger)
		default:
			replyWithErrorLogging(
				interactionReplier(s, i),
				"Unknown button interaction."+internalError,
				logger,
			)
		}

		return
	}

	replyWithErrorLogging(
		interactionReplier(s, i),
		"Unknown button interaction."+internalError,
		logger,
	)

}

func (h *EventHandler) handleVerifyButton(
	d *VerificationDialog,
	s *discordgo.Session,
	i *discordgo.Interaction,
	l *zap.Logger,
) {

	err := h.EventScoreService.Verify(d.SID)
	if err != nil {
		l.Error("could not verify score", zap.Error(err))
		replyWithErrorLogging(interactionReplier(s, i), "Could not verify score."+internalError, l)
		return
	}

	d.Verified = true
	d.VerifiedBy = formatMember(i.Member)

	l.Debug("interaction in verify button", zap.Any("interaction", i))

	err = s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{d.ToEmbed()},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Next",
							Style:    discordgo.PrimaryButton,
							CustomID: scoreNextButton,
						},
					},
				},
			},
		},
	})

	if err != nil {
		l.Error("could not edit interaction response", zap.Error(err))
		replyWithErrorLogging(interactionReplier(s, i), "Could not update embed."+internalError, l)
	}
}

func (h *EventHandler) handleRejectButton(
	d *VerificationDialog,
	s *discordgo.Session,
	i *discordgo.Interaction,
	l *zap.Logger,
) {
	err := h.EventScoreService.Verify(d.SID)
	if err != nil {
		l.Error("could not verify score", zap.Error(err))
		replyWithErrorLogging(interactionReplier(s, i), "Could not verify score."+internalError, l)
		return
	}

	err = s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				d.ToEmbed(),
				{
					Title:       "Instruction",
					Description: "Please use the update command to give the submission a new score or remove this entry",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name: "Template",
							Value: fmt.Sprintf(
								"```\n/events update-score submission-id: %s new-score: <new score>\n```",
								d.SID,
							),
						},
					},
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Next",
							Style:    discordgo.PrimaryButton,
							CustomID: scoreNextButton,
						},
						discordgo.Button{
							Label:    "Remove",
							Style:    discordgo.DangerButton,
							CustomID: scoreRemoveButton,
						},
					},
				},
			},
		},
	})

	if err != nil {
		l.Error("could not edit interaction response", zap.Error(err))
		replyWithErrorLogging(interactionReplier(s, i), "Could not update embed."+internalError, l)
	}
}

func (h *EventHandler) handleNextButton(
	d *VerificationDialog,
	s *discordgo.Session,
	i *discordgo.Interaction,
	l *zap.Logger,
) {
	record, err := h.EventScoreService.GetOneUnverifiedForEvent(d.EID)
	if err != nil {
		if scores.AsErrNoRecord(err) {
			replyWithErrorLogging(
				interactionReplier(s, i),
				":tada: There are no pending submissions to be verified",
				l,
			)
			return
		}
		replyWithErrorLogging(
			interactionReplier(s, i),
			"Sorry, something's borked."+internalError,
			l,
		)
		return
	}

	member, err := s.GuildMember(i.GuildID, record.UID)
	if err != nil {
		l.Error("could not fetch user information", zap.Error(err))
		replyWithErrorLogging(
			interactionReplier(s, i),
			"Something went wrong."+internalError,
			l,
		)
		return
	}

	verifDialog := &VerificationDialog{
		UserDisplay: formatMember(member),
		SID:         record.ID,
		IGN:         "`" + record.IGN + "`",
		Score:       fmt.Sprintf("%v", record.Score),
		URL:         record.Proof,
		EID:         d.EID,
		EventName:   d.EventName,
	}

	err = s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				verifDialog.ToEmbed(),
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Verify",
							Style:    discordgo.SuccessButton,
							CustomID: scoreVerificationBotton,
						},
						discordgo.Button{
							Label:    "Reject",
							Style:    discordgo.DangerButton,
							CustomID: scoreRejectButton,
						},
					},
				},
			},
		},
	})

	if err != nil {
		l.Error("could not respond to interaction", zap.Error(err))
	}
}

func (h *EventHandler) handleRemoveButton(
	d *VerificationDialog,
	s *discordgo.Session,
	i *discordgo.Interaction,
	l *zap.Logger,
) {
	err := h.EventScoreService.DeleteScore(d.SID)
	if err != nil {
		l.Error("could not delete score", zap.Error(err))
		replyWithErrorLogging(interactionReplier(s, i), "Error deleting score."+internalError, l)
		return
	}

	h.handleNextButton(d, s, i, l)
}
