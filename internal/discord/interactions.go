package discord

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/2785/warframe-assistant/internal/meta"
	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
)

func (h *EventHandler) HandleInteractionsCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := h.InteractionCreateHandlers[i.Data.Name]; ok {
		h(s, i)
		return
	}

	h.Logger.Warn("unknown slash command", zap.String("command", i.Data.Name))
}

func (h *EventHandler) RegisterInteractionCreateHandlers(s *discordgo.Session) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "test",
			Description: "test slash command for warframe assistant",
		},
		{
			Name:        "ign",
			Description: "Commands regarding IGN management",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "register",
					Description: "Associate your IGN with your discord user ID in the bot",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "ign",
							Description: "Your Warframe in game name",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "purge",
					Description: "Remove your IGN and all associated records from the bot",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "update",
					Description: "Update your IGN associated with your discord user ID in the bot",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "ign",
							Description: "Your Warframe in game name",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "help",
			Description: "Display information regarding what the bot does / how to use the bot",
		},
		{
			Name:        "events",
			Description: "Event information and management",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List events in this guild",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "print-all",
							Description: "If all events should be printed",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "create",
					Description: "Create a new event",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the new event",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "type",
							Description: fmt.Sprintf("Type of the new event, supported types: %s", strings.Join(supportedEventTypes, ", ")),
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "end-date",
							Description: "End date of the event, in the format of `Jan 2, 2006 at 3:04pm (MST)`",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "start-date",
							Description: "Start date of the event, in the format of `Jan 2, 2006 at 3:04pm (MST)`",
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "active",
							Description: "If the event is created as an active event, defaults to yes",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "join",
					Description: "Join the active event if ID unspecified, else join the specified event",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "event-id",
							Description: "The UUID of the event (this flow will be updated later)",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "purge-participation",
					Description: "Purge all participation records regarding the active event",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "event-id",
							Description: "The UUID of the event (this flow will be updated later)",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "bail",
					Description: "Quit the active event if ID unspecified, else quit the specified event",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "event-id",
							Description: "The UUID of the event (this flow will be updated later)",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "activate",
					Description: "Activate a specified event by ID",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "event-id",
							Description: "The UUID of the event",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "deactivate",
					Description: "Deactivate a specified event by ID",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "event-id",
							Description: "The UUID of the event",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list-participant",
					Description: "List participants of an event",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "event-id",
							Description: "The UUID of the event",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "progress",
					Description: "Print the progress of current events",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "event-id",
							Description: "The UUID of the event, prints all if omitted",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "verify",
					Description: "Trigger the submission verification work flow",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "event-id",
							Description: "The UUID of the event",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "update-score",
					Description: "Update the score of a submission and verify the submission",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "submission-id",
							Description: "The UUID of the submission",
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "new-score",
							Description: "New score for the submission",
						},
					},
				},
			},
		},
	}

	handlers := map[string]func(*discordgo.Session, *discordgo.InteractionCreate){
		"test":   h.handleTest,
		"ign":    h.handleIGN,
		"events": h.handleEvents,
		"help":   h.handleHelp,
	}

	h.Commands = commands
	h.InteractionCreateHandlers = handlers

	// for _, v := range commands {
	// 	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
	// 	if err != nil {
	// 		h.Logger.Error("could not register application command", zap.String("command", v.Name), zap.Error(err))
	// 		return err
	// 	}
	// }

	return nil
}

func interactionReplier(s *discordgo.Session, i *discordgo.Interaction) MessageReplier {
	return func(msg string) error {
		return s.InteractionRespond(i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: msg,
			},
		})
	}
}

func (h *EventHandler) handleTest(s *discordgo.Session, i *discordgo.InteractionCreate) {
	replyWithErrorLogging(interactionReplier(s, i.Interaction), "Hello from warframe assistant", h.Logger.With(zap.String("handler", "test-handler")))
}

func (h *EventHandler) handleIGN(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if len(i.Data.Options) < 1 {
		h.interactionRespondWithErrorLogging(s, i.Interaction, "No subcommand found")
		return
	}

	switch i.Data.Options[0].Name {
	case "register":
		if len(i.Data.Options[0].Options) < 1 {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "No ign input found")
			return
		}

		ign, ok := i.Data.Options[0].Options[0].Value.(string)
		if !ok {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "ign must be a string")
			return
		}
		if ign == "" {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "ign cannot be empty")
			return
		}

		err := h.MetadataService.CreateIGN(i.Member.User.ID, ign)
		if err != nil {
			dupErr := &meta.ErrDuplicateEntry{}
			if errors.As(err, &dupErr) {
				existing, err := h.MetadataService.GetIGN(i.Member.User.ID)
				if err != nil {
					h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not add the ign due to a dup error, yet could not retrieve existing ign, something is borked, please try again later or contact bot maintainer for help")
					return
				}
				h.interactionRespondWithErrorLogging(s, i.Interaction, fmt.Sprintf("Your discord user already has an associated IGN, `%s`. Please use the update command if you would like to change it", existing))
				return
			}
			h.Logger.Error("could not add ign", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not add the ign, something is borked."+internalError)
			return
		}

		h.interactionRespondWithErrorLogging(s, i.Interaction, fmt.Sprintf("Successfully associated your discord user with the IGN `%s`", ign))
		return

	case "update":
		oldIGN, ok := h.mustHaveIGNRegistered(i.Member.User.ID, interactionReplier(s, i.Interaction))
		if !ok {
			return
		}

		if len(i.Data.Options[0].Options) < 1 {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "No ign input found")
			return
		}

		ign, ok := i.Data.Options[0].Options[0].Value.(string)
		if !ok {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "ign must be a string")
			return
		}
		if ign == "" {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "ign cannot be empty")
			return
		}

		err := h.MetadataService.UpdateIGN(i.Member.User.ID, ign)
		if err != nil {
			h.Logger.Error("could not update ign", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not update the ign, something is borked."+internalError)
			return
		}

		h.interactionRespondWithErrorLogging(s, i.Interaction, fmt.Sprintf("Successfully updated the IGN associated your discord user from `%s` to `%s`", oldIGN, ign))
		return

	case "purge":
		_, ok := h.mustHaveIGNRegistered(i.Member.User.ID, interactionReplier(s, i.Interaction))
		if !ok {
			return
		}
		err := h.MetadataService.DeleteRelation(i.Member.User.ID)
		if err != nil {
			h.Logger.Error("could not purge user ign", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not purge the ign registration, something is borked."+internalError)
			return
		}
		h.interactionRespondWithErrorLogging(s, i.Interaction, "Successfully purged ign relation with all associated data")
		return

	default:
		h.interactionRespondWithErrorLogging(s, i.Interaction, "Unknown subcommand")
		return
	}

}

func (h *EventHandler) handleEvents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if len(i.Data.Options) < 1 {
		h.interactionRespondWithErrorLogging(s, i.Interaction, "No subcommand found")
		return
	}

	subCmd := i.Data.Options[0]

	roleRequirement, err := h.MetadataService.GetRoleRequirementForGuild(string(manageEventDialog), i.GuildID)

	if err != nil {
		h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
		return
	}

	switch subCmd.Name {
	case "list":
		op := bindOptions(subCmd.Options)
		all, ok := op["print-all"].(bool)
		if !ok {
			all = false
		}

		var events []*meta.Event
		var err error

		if all {
			events, err = h.MetadataService.ListEventsForGuild(i.GuildID)
		} else {
			events, err = h.MetadataService.ListActiveEventsForGuild(i.GuildID)
		}

		if err != nil {
			h.Logger.Error("could not list all events", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, fmt.Sprintf("Could not list events."+internalError))
			return
		}

		if len(events) == 0 {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "There are currently no events!")
			return
		}

		embeds := make([]*discordgo.MessageEmbed, len(events))

		for i, v := range events {
			embeds[i] = &discordgo.MessageEmbed{
				Title: v.Name,
				Fields: []*discordgo.MessageEmbedField{
					{Name: "ID", Value: v.ID},
					{Name: "Start Time", Value: formatTime(v.Begin)},
					{Name: "End Time", Value: formatTime(v.End)},
					{Name: "Status", Value: func() string {
						if v.Active {
							return "Active"
						} else {
							return "Inactive"
						}
					}()},
					{Name: "Type", Value: v.EventType},
				},
			}
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Embeds: embeds,
			},
		})

		if err != nil {
			h.Logger.Error("could not send response to intraction", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not list events."+internalError)
			return
		}

	case "create":
		if !h.mustHaveRoleWithID(i.Member.User.ID, roleRequirement, i.GuildID, interactionReplier(s, i.Interaction), s) {
			return
		}

		op := bindOptions(subCmd.Options)

		name, nameOk := op["name"].(string)
		eType, typeOk := op["type"].(string)
		endDates, endDateOk := op["end-date"].(string)

		if !(nameOk && typeOk && endDateOk) {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Name, type, and end date must be provided")
			return
		}

		if !funk.Contains(supportedEventTypes, eType) {
			h.interactionRespondWithErrorLogging(s, i.Interaction, fmt.Sprintf("Sorry, only the following event types are currently supported: %s", strings.Join(supportedEventTypes, ", ")))
			return
		}

		formatStr := "Jan 2, 2006 at 3:04pm (MST)"
		endDate, err := time.Parse(formatStr, endDates)
		if err != nil {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "end date format must be of "+formatStr)
			return
		}

		startDates, startDateOk := op["start-date"].(string)
		active, activeOk := op["active"].(bool)
		if !activeOk {
			active = true
		}
		var startDate time.Time
		if startDateOk {
			var err error
			startDate, err = time.Parse(formatStr, startDates)
			if err != nil {
				h.interactionRespondWithErrorLogging(s, i.Interaction, "start date format must be of "+formatStr)
				return
			}
		} else {
			startDate = time.Now()
		}

		eid, err := h.MetadataService.CreateEvent(name, eType, startDate, endDate, i.GuildID, active)
		if err != nil {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not create event."+internalError)
			return
		}

		h.interactionRespondWithErrorLogging(s, i.Interaction, fmt.Sprintf("Successfully created event with ID '%s'", eid))
		return

	case "join":
		op := bindOptions(subCmd.Options)
		eid, ok := op["event-id"].(string)
		if !ok {
			eid = ""
		}

		// must have the user's IGN on file to join event
		if _, ok := h.mustHaveIGNRegistered(i.Member.User.ID, interactionReplier(s, i.Interaction)); !ok {
			return
		}

		logger := h.Logger.With(WithGuildID(i.GuildID), WithUserID(i.Member.User.ID))

		// if id is not specified we check if there's only one active event
		if eid == "" {
			eid, ok = h.mustGetOneActiveEventIDForGuild(i.GuildID, interactionReplier(s, i.Interaction))
			if !ok {
				return
			}
		}

		pid, in, err := h.MetadataService.GetParticipation(i.Member.User.ID, eid)
		if err != nil {
			if meta.AsErrNoRecord(err) {
				_, err := h.MetadataService.AddParticipation(i.Member.User.ID, eid, true)
				if err != nil {
					logger.Error("could not add participant to event", zap.Error(err), WithEventID(eid))
					h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not join the event."+internalError)
					return
				}
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Successfully joined the event")
				return
			}

			logger.Error("could not check if user is already in event", zap.Error(err), WithEventID(eid))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not join the event."+internalError)
			return
		}

		if in {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "You are already in this event")
			return
		} else {
			err = h.MetadataService.SetParticipation(pid, true)
			if err != nil {
				logger.Error("could not update participation", zap.Error(err), WithEventID(eid))
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong while trying to update participation status."+internalError)
				return
			}

			h.interactionRespondWithErrorLogging(s, i.Interaction, "Successfully updated your participation status")
		}

		return

	case "bail":
		op := bindOptions(subCmd.Options)
		eid, ok := op["event-id"].(string)
		if !ok {
			eid = ""
		}

		// must have the user's IGN on file to join event
		if _, ok := h.mustHaveIGNRegistered(i.Member.User.ID, interactionReplier(s, i.Interaction)); !ok {
			return
		}

		logger := h.Logger.With(WithGuildID(i.GuildID), WithUserID(i.Member.User.ID))

		// if id is not specified we check if there's only one active event
		if eid == "" {
			eid, ok = h.mustGetOneActiveEventIDForGuild(i.GuildID, interactionReplier(s, i.Interaction))
			if !ok {
				return
			}
		}

		pid, in, err := h.MetadataService.GetParticipation(i.Member.User.ID, eid)
		if err != nil {
			if meta.AsErrNoRecord(err) {
				_, err := h.MetadataService.AddParticipation(i.Member.User.ID, eid, false)
				if err != nil {
					logger.Error("could not add participant status to event", zap.Error(err), WithEventID(eid))
					h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not bail the event."+internalError)
					return
				}

				h.interactionRespondWithErrorLogging(s, i.Interaction, "Successfully bailed the event")
				return
			}

			logger.Error("could not check if user is already in event", zap.Error(err), WithEventID(eid))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not join the event."+internalError)
			return
		}

		if !in {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "You have already bailed this event")
			return
		} else {
			err = h.MetadataService.SetParticipation(pid, false)
			if err != nil {
				logger.Error("could not update participation", zap.Error(err), WithEventID(eid))
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong while trying to update participation status."+internalError)
				return
			}

			h.interactionRespondWithErrorLogging(s, i.Interaction, "Successfully updated your participation status")
		}

	case "purge-participation":
		op := bindOptions(subCmd.Options)
		eid, ok := op["event-id"].(string)
		if !ok {
			eid = ""
		}

		// must have the user's IGN on file to join event
		if _, ok := h.mustHaveIGNRegistered(i.Member.User.ID, interactionReplier(s, i.Interaction)); !ok {
			return
		}

		// if id is not specified we check if there's only one active event
		if eid == "" {
			eid, ok = h.mustGetOneActiveEventIDForGuild(i.GuildID, interactionReplier(s, i.Interaction))
			if !ok {
				return
			}
		}

		logger := h.Logger.With(WithGuildID(i.GuildID), WithUserID(i.Member.User.ID), WithEventID(eid))

		pid, _, err := h.MetadataService.GetParticipation(i.Member.User.ID, eid)
		if err != nil {
			if meta.AsErrNoRecord(err) {
				replyWithErrorLogging(interactionReplier(s, i.Interaction), "You are not registered in the event, there's nothing to purge", logger)
				return
			}
			replyWithErrorLogging(interactionReplier(s, i.Interaction), "Could not check if you are in the event."+internalError, logger)
			return
		}

		err = h.MetadataService.DeleteParticipation(pid)
		if err != nil {
			replyWithErrorLogging(interactionReplier(s, i.Interaction), "Could not delete your participation of the event."+internalError, logger)
		}

		replyWithErrorLogging(interactionReplier(s, i.Interaction), "Successfully deleted your participation record in the event", logger)

		return

	case "activate":
		if !h.mustHaveRoleWithID(i.Member.User.ID, roleRequirement, i.GuildID, interactionReplier(s, i.Interaction), s) {
			return
		}

		op := bindOptions(subCmd.Options)
		id, ok := op["event-id"].(string)
		if !ok || id == "" {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "You must supply an event ID to activate")
		}

		err := h.MetadataService.SetEventStatus(id, true)
		if err != nil {
			h.Logger.Error("could not set event status", zap.Error(err))
			return
		}

		h.interactionRespondWithErrorLogging(s, i.Interaction, "Successfully activated event")
		return

	case "deactivate":
		if !h.mustHaveRoleWithID(i.Member.User.ID, roleRequirement, i.GuildID, interactionReplier(s, i.Interaction), s) {
			return
		}

		op := bindOptions(subCmd.Options)
		id, ok := op["event-id"].(string)
		if !ok || id == "" {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "You must supply an event ID to activate")
		}

		err := h.MetadataService.SetEventStatus(id, false)
		if err != nil {
			h.Logger.Error("could not set event status", zap.Error(err))
			return
		}

		h.interactionRespondWithErrorLogging(s, i.Interaction, "Successfully deactivated event")
		return

	case "list-participant":
		op := bindOptions(subCmd.Options)
		eid, ok := op["event-id"].(string)
		if !ok {
			eid = ""
		}

		// if id is not specified we check if there's only one active event
		if eid == "" {
			eid, ok = h.mustGetOneActiveEventIDForGuild(i.GuildID, interactionReplier(s, i.Interaction))
			if !ok {
				return
			}
		}

		logger := h.Logger.With(WithGuildID(i.GuildID), WithEventID(eid), WithCommand("list-participant"))

		usersIn, usersOut, err := h.MetadataService.ListUserForEvent(eid)
		if err != nil {
			logger.Error("could not list users is in event", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
			return
		}

		event, err := h.MetadataService.GetEvent(eid)

		if err != nil {
			logger.Error("could not get event by ID", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
			return
		}

		usersInDisp, usersOutDisp := make([]string, 0, len(usersIn)), make([]string, 0, len(usersOut))
		for k, v := range usersIn {
			member, err := s.GuildMember(i.GuildID, k)
			if err != nil {
				logger.Error("could not fetch user information", zap.Error(err))
				replyWithErrorLogging(interactionReplier(s, i.Interaction), "Error fetching user information."+internalError, logger)
				return
			}
			usersInDisp = append(usersInDisp, fmt.Sprintf("%s (`%s`)", func() string {
				if member.Nick != "" {
					return member.Nick
				}
				return member.User.Username + "#" + member.User.Discriminator
			}(), v))
		}

		for k, v := range usersOut {
			member, err := s.GuildMember(i.GuildID, k)
			if err != nil {
				logger.Error("could not fetch user information", zap.Error(err))
				replyWithErrorLogging(interactionReplier(s, i.Interaction), "Error fetching user information."+internalError, logger)
				return
			}
			usersOutDisp = append(usersOutDisp, fmt.Sprintf("%s#%s (`%s`)", member.User.Username, member.User.Discriminator, v))
		}

		embeds := []*discordgo.MessageEmbedField{}
		if len(usersInDisp) > 0 {
			embeds = append(embeds, &discordgo.MessageEmbedField{Name: "Joined", Value: strings.Join(usersInDisp, "\n")})
		}

		if len(usersOutDisp) > 0 {
			embeds = append(embeds, &discordgo.MessageEmbedField{Name: "Bailed", Value: strings.Join(usersOutDisp, "\n")})
		}

		if len(embeds) == 0 {
			embeds = append(embeds, &discordgo.MessageEmbedField{Name: "Nothing", Value: "There's absolutely nothing in this event"})
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:  "Event: " + event.Name,
						Fields: embeds,
					},
				},
			},
		})

		if err != nil {
			logger.Error("could not send embed", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
			return
		}

	case "progress":
		op := bindOptions(subCmd.Options)
		eid, ok := op["event-id"].(string)
		if !ok {
			eid = ""
		}

		// if id is not specified we check if there's only one active event
		if eid == "" {
			eid, ok = h.mustGetOneActiveEventIDForGuild(i.GuildID, interactionReplier(s, i.Interaction))
			if !ok {
				return
			}
		}

		logger := h.Logger.With(WithGuildID(i.GuildID), WithEventID(eid), WithCommand("list-participant"))
		plainTextReplier := interactionReplier(s, i.Interaction)

		event, err := h.MetadataService.GetEvent(eid)
		if err != nil {
			logger.Error("error fetching event information", zap.Error(err))
			replyWithErrorLogging(plainTextReplier, "Could not fetch event information."+internalError, logger)
			return
		}

		switch event.EventType {
		case eventTypeTournament:
			replyWithErrorLogging(plainTextReplier, "Ay mate progress display for tournaments ain't implemented yet", logger)
			return
		case eventTypeScoreCampaign:
			leaderboard, err := h.EventScoreService.MakeReportScoreSum(eid)
			if err != nil {
				replyWithErrorLogging(plainTextReplier, "Could not fetch event information."+internalError, logger)
				return
			}

			if len(leaderboard) == 0 {
				replyWithErrorLogging(plainTextReplier, "There's no submissions in this event yet!", logger)
				return
			}

			fields := make([]string, len(leaderboard))

			for i, v := range leaderboard {
				user, err := s.GuildMember(event.GID, v.UID)
				if err != nil {
					replyWithErrorLogging(plainTextReplier, "Could not fetch user information."+internalError, logger)
					return
				}
				fields[i] = fmt.Sprintf("#%v - %s (`%s`) - %v points", i+1, func() string {
					if user.Nick != "" {
						return user.Nick
					}
					return user.User.Username + "#" + user.User.Discriminator
				}(), v.IGN, v.Score)
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: event.Name + " (Accumulative)",
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:  "Leaderboard",
									Value: strings.Join(fields, "\n"),
								},
							},
						},
					},
				},
			})

			if err != nil {
				logger.Error("could not send embeds", zap.Error(err))
				replyWithErrorLogging(plainTextReplier, "Error fetching event leadboard."+internalError, logger)
			}

			return

		case eventTypeScoreLeaderboard:
			leaderboard, err := h.EventScoreService.MakeReportScoreTop(eid)
			if err != nil {
				replyWithErrorLogging(plainTextReplier, "Could not fetch event information."+internalError, logger)
				return
			}

			if len(leaderboard) == 0 {
				replyWithErrorLogging(plainTextReplier, "There's no submissions in this event yet!", logger)
				return
			}

			fields := make([]string, len(leaderboard))

			for i, v := range leaderboard {
				user, err := s.GuildMember(event.GID, v.UID)
				if err != nil {
					replyWithErrorLogging(plainTextReplier, "Could not fetch user information."+internalError, logger)
					return
				}
				fields[i] = fmt.Sprintf("#%v - %s (`%s`) - %v points", i+1, func() string {
					if user.Nick != "" {
						return user.Nick
					}
					return user.User.Username + "#" + user.User.Discriminator
				}(), v.IGN, v.Score)
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: event.Name + " (Only best score counts)",
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:  "Leaderboard",
									Value: strings.Join(fields, "\n"),
								},
							},
						},
					},
				},
			})

			if err != nil {
				logger.Error("could not send embeds", zap.Error(err))
				replyWithErrorLogging(plainTextReplier, "Error fetching event leadboard."+internalError, logger)
			}
			return
		default:
			replyWithErrorLogging(plainTextReplier, "unknown event type "+event.EventType, logger)
			return
		}
	case "verify":
		if !h.mustHaveRoleWithID(i.Member.User.ID, roleRequirement, i.GuildID, interactionReplier(s, i.Interaction), s) {
			return
		}

		op := bindOptions(subCmd.Options)
		eid, ok := op["event-id"].(string)
		if !ok {
			eid = ""
		}

		// if id is not specified we check if there's only one active event
		if eid == "" {
			eid, ok = h.mustGetOneActiveEventIDForGuild(i.GuildID, interactionReplier(s, i.Interaction))
			if !ok {
				return
			}
		}

		replyWithErrorLogging(interactionReplier(s, i.Interaction), "Gotcha", h.Logger.With(WithCommand("events verify"), WithEventID(eid), WithUserID(i.Member.User.ID)))

		h.handleGetOneUnverifiedChannel(s, i.GuildID, i.ChannelID, eid)
		return
	default:
		h.interactionRespondWithErrorLogging(s, i.Interaction, "unknown subcommand")
	}
}

func (h *EventHandler) handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Warframe Assistant User Manual",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "General Information",
							Value: "The bot is developed to assist in running warframe clan events and related utilities in discord, the bot is under _active-ish_ development and attempts will be made to make this guide up to date",
						},
						{
							Name: "IGN Management",
							Value: strings.Join([]string{
								"You can register your in game name with the bot, this will be required in order to participate in events.",
								"`/ign register` - associate your IGN with the discord user ID",
								"`/ign purge` - remove the association from the database, this will purge all event scores",
								"`/ign update` - updates the ign associated with your account",
							}, "\n"),
						},
						{
							Name: "Event Management - part 1",
							Value: strings.Join([]string{
								"Event related utilities come under the events command, two types of events are currently supported",
								"`scoreboard-campaign` - where participants claim scores with screenshot proofs and ones with the highest accumulated score wins",
								"`scoreboard-leaderboard` - where participants claim scores with screenshot proofs and ones with the top single score wins",
								"`tournament` - single elimination random matchup pvp tournament",
								"`?!submit <score> event: <event-id>` - submit a screenshot to claim a score, event ID can be omitted if there's only one active event",
								"`/events list` - list active events, optionally pass argument to list all events",
							}, "\n"),
						},
						{
							Name: "Event Management - part 2",
							Value: strings.Join([]string{
								"`/events join` - join the event specified with the event ID, or join the only active event",
								"`/events bail` - leave an event specified with the event ID, or the only active event",
								"`/events purge-participation` - nukes all record of you ever doing anything with this event",
								"`/events list-participant` - list the participants of the event specified with the event ID, or the only active event",
								"`/events progress` - checks the progress of event specified by the event ID, or the only active event",
								"mod only: `/events create` - creates an event, if start time is unspecified, current time will be used",
								"mod only: `/events activate` - activate an event by ID",
								"mod only: `/events deactivate` - deactivate an event by ID",
								"mod only: `/events verify` - triggers the verification workflow",
							}, "\n"),
						},
					},
				},
			},
		},
	})

	if err != nil {
		h.Logger.Error("could not send help message", zap.Error(err))
		h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
	}
}

func (h *EventHandler) interactionRespondWithErrorLogging(s *discordgo.Session, i *discordgo.Interaction, msg string) {
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: msg,
		},
	})

	if err != nil {
		h.Logger.Error("could not respond to interaction", zap.Error(err))
	}
}

func bindOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]interface{} {
	out := make(map[string]interface{}, len(options))
	for _, v := range options {
		out[v.Name] = v.Value
	}

	return out
}

func formatTime(d time.Time) string {
	t := d.Format("Mon Jan 2 15:04 MST")
	now := time.Now()

	if d.Before(now) {
		t += fmt.Sprintf(", %s ago", durafmt.Parse(time.Since(d)).LimitFirstN(2))
	} else {
		t += fmt.Sprintf(", in %s", durafmt.Parse(d.Sub(now)).LimitFirstN(2))
	}

	return t
}
