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
		// {
		// 	Name:        "test",
		// 	Description: "test slash command for warframe assistant",
		// },
		// {
		// 	Name:        "ign",
		// 	Description: "Commands regarding IGN management",
		// 	Options: []*discordgo.ApplicationCommandOption{
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "register",
		// 			Description: "Associate your IGN with your discord user ID in the bot",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "ign",
		// 					Description: "Your Warframe in game name",
		// 					Required:    true,
		// 				},
		// 			},
		// 		},
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "purge",
		// 			Description: "Remove your IGN and all associated records from the bot",
		// 		},
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "update",
		// 			Description: "Update your IGN associated with your discord user ID in the bot",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "ign",
		// 					Description: "Your Warframe in game name",
		// 					Required:    true,
		// 				},
		// 			},
		// ok :=
		// 	},
		// },
		{
			Name:        "help",
			Description: "Display information regarding what the bot does / how to use the bot",
		},
		// {
		// 	Name:        "events",
		// 	Description: "Event information and management",
		// 	Options: []*discordgo.ApplicationCommandOption{
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "list",
		// 			Description: "List events in this guild",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionBoolean,
		// 					Name:        "print-all",
		// 					Description: "If all events should be printed",
		// 				},
		// 			},
		// 		},
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "create",
		// 			Description: "Create a new event",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "name",
		// 					Description: "Name of the new event",
		// 					Required:    true,
		// 				},
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "type",
		// 					Description: "Type of the new event, for now supports `scoreboard-campaign` and `tournament`",
		// 					Required:    true,
		// 				},
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "end-date",
		// 					Description: "End date of the event, in the format of `Jan 2, 2006 at 3:04pm (MST)`",
		// 					Required:    true,
		// 				},
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "start-date",
		// 					Description: "Start date of the event, in the format of `Jan 2, 2006 at 3:04pm (MST)`",
		// 				},
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionBoolean,
		// 					Name:        "active",
		// 					Description: "If the event is created as an active event, defaults to yes",
		// 				},
		// 			},
		// 		},
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "join",
		// 			Description: "Join the active event if ID unspecified, else join the specified event",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "event-id",
		// 					Description: "The UUID of the event (this flow will be updated later)",
		// 				},
		// 			},
		// 		},
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "bail",
		// 			Description: "Quit the active event if ID unspecified, else quit the specified event",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "event-id",
		// 					Description: "The UUID of the event (this flow will be updated later)",
		// 				},
		// 			},
		// 		},
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "activate",
		// 			Description: "Activate a specified event by ID",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "event-id",
		// 					Description: "The UUID of the event",
		// 					Required:    true,
		// 				},
		// 			},
		// 		},
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "deactivate",
		// 			Description: "Deactivate a specified event by ID",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "event-id",
		// 					Description: "The UUID of the event",
		// 					Required:    true,
		// 				},
		// 			},
		// 		},
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "list-participant",
		// 			Description: "List participants of an event",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "event-id",
		// 					Description: "The UUID of the event",
		// 				},
		// 			},
		// 		},
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "progress",
		// 			Description: "Print the progress of current events",
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Name:        "event-id",
		// 					Description: "The UUID of the event, prints all if omitted",
		// 				},
		// 			},
		// 		},
		// 	},
		// },
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

func (h *EventHandler) handleTest(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: "hello from warframe assistant",
		},
	})

	if err != nil {
		h.Logger.Error("could not send interaction response", zap.Error(err))
		return
	}
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

		h.interactionRespondWithErrorLogging(s, i.Interaction, fmt.Sprintf("Successfully updated the IGN associated your discord user to `%s`", ign))
		return
	case "purge":
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
		if !h.ensureUserHasRoleInteraction(s, i.Member.User.ID, i.GuildID, i.ChannelID, roleRequirement, string(manageEventDialog), i.Interaction) {
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

		if !funk.Contains([]string{"scoreboard-campaign", "tournament"}, eType) {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Sorry, only 'tournament' and 'scoreboard-campaign' are supported at the moment")
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
		id, ok := op["event-id"].(string)
		if !ok {
			id = ""
		}

		// if we don't have the user's IGN we ask for the IGn
		_, err := h.MetadataService.GetIGN(i.Member.User.ID)
		if err != nil {
			nr := &meta.ErrNoRecord{}
			if errors.As(err, &nr) {
				h.interactionRespondWithErrorLogging(s, i.Interaction, "You must have your IGN registered with the bot to join an event, please use the '/ign register` slash command")
				return
			}
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong while checking for IGN."+internalError)
			return
		}

		// if id is specified we do that id
		if id != "" {
			_, err := h.MetadataService.AddParticipation(i.Member.User.ID, id)
			if err != nil {
				h.Logger.Error("could not add participant to event", zap.Error(err), zap.String("gid", i.GuildID), zap.String("uid", i.Member.User.ID), zap.String("eid", id))
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not join the event."+internalError)
				return
			}
		}

		// if id is not specified, we check how many active events there are
		events, err := h.MetadataService.ListActiveEventsForGuild(i.GuildID)
		if err != nil {
			h.Logger.Error("could not list events", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not list active events to join."+internalError)
		}

		if len(events) == 0 {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "There's no active events to join.")
			return
		} else if len(events) > 1 {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "There are multiple active events, please specify one by ID.")
			return
		}

		_, err = h.MetadataService.AddParticipation(i.Member.User.ID, events[0].ID)
		if err != nil {
			dup := &meta.ErrDuplicateEntry{}
			if errors.As(err, &dup) {
				h.interactionRespondWithErrorLogging(s, i.Interaction, "You are already in this event")
				return
			}
			h.Logger.Error("could not add participation", zap.Error(err), zap.String("gid", i.GuildID), zap.String("uid", i.Member.User.ID), zap.String("eid", events[0].ID))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Found one active event to join but encountered an error."+internalError)
			return
		}

		h.interactionRespondWithErrorLogging(s, i.Interaction, fmt.Sprintf("Successfully added you to the event '%s'", events[0].Name))
		return

	case "bail":
		op := bindOptions(subCmd.Options)
		id, ok := op["event-id"].(string)
		if !ok {
			id = ""
		}

		if id != "" {
			in, err := h.MetadataService.UserInEvent(i.Member.User.ID, id)
			if err != nil {
				h.Logger.Error("could not check if user is in event", zap.Error(err))
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
				return
			}
			if !in {
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Ya ain't in that event mate no need to bail")
				return
			}
			err = h.MetadataService.DeleteParticipationByUserAndEvent(i.Member.User.ID, id)
			if err != nil {
				h.Logger.Error("could not bail user", zap.Error(err))
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
				return
			}

			h.interactionRespondWithErrorLogging(s, i.Interaction, "Successfully bailed from event")
			return
		}

		events, err := h.MetadataService.ListActiveEventsForGuild(i.GuildID)
		if err != nil {
			h.Logger.Error("could not list events", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not list active events to bail from."+internalError)
		}

		if len(events) == 0 {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "There's no active events to bail from.")
			return
		} else if len(events) > 1 {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "There are multiple active events, please specify one by ID.")
			return
		}

		in, err := h.MetadataService.UserInEvent(i.Member.User.ID, events[0].ID)
		if err != nil {
			h.Logger.Error("could not check if user is in event", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
			return
		}
		if !in {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Ya ain't in that event mate no need to bail")
			return
		}

		err = h.MetadataService.DeleteParticipationByUserAndEvent(i.Member.User.ID, events[0].ID)
		if err != nil {
			h.Logger.Error("could not bail user", zap.Error(err))
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
			return
		}

		h.interactionRespondWithErrorLogging(s, i.Interaction, "Successfully bailed from event")
		return

	case "activate":
		if !h.ensureUserHasRoleInteraction(s, i.Member.User.ID, i.GuildID, i.ChannelID, roleRequirement, string(manageEventDialog), i.Interaction) {
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
		if !h.ensureUserHasRoleInteraction(s, i.Member.User.ID, i.GuildID, i.ChannelID, roleRequirement, string(manageEventDialog), i.Interaction) {
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
		id, ok := op["event-id"].(string)
		if !ok {
			id = ""
		}

		var users map[string]string
		var event *meta.Event

		if id != "" {
			var err error
			users, err = h.MetadataService.ListUserForEvent(id)
			if err != nil {
				h.Logger.Error("could not list users is in event", zap.Error(err))
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
				return
			}

			event, err = h.MetadataService.GetEvent(id)
			if err != nil {
				h.Logger.Error("could not get event by ID", zap.Error(err))
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
				return
			}
		} else {
			events, err := h.MetadataService.ListActiveEventsForGuild(i.GuildID)
			if err != nil {
				h.Logger.Error("could not list events", zap.Error(err))
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Could not list active events."+internalError)
			}

			if len(events) == 0 {
				h.interactionRespondWithErrorLogging(s, i.Interaction, "There are no active events.")
				return
			} else if len(events) > 1 {
				h.interactionRespondWithErrorLogging(s, i.Interaction, "There are multiple active events, please specify one by ID.")
				return
			}

			users, err = h.MetadataService.ListUserForEvent(events[0].ID)
			if err != nil {
				h.Logger.Error("could not list users is in event", zap.Error(err))
				h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong."+internalError)
				return
			}

			event = events[0]
		}

		usersDisplay := make([]string, 0, len(users))
		for k, v := range users {
			member, err := s.GuildMember(i.GuildID, k)
			if err != nil {
				h.Logger.Error("could not fetch user information", zap.Error(err))
				return
			}
			usersDisplay = append(usersDisplay, fmt.Sprintf("%s#%s (`%s`)", member.User.Username, member.User.Discriminator, v))
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title: "Event: " + event.Name,
						Fields: []*discordgo.MessageEmbedField{
							{Name: "ID", Value: event.ID},
							{Name: "Participants", Value: strings.Join(usersDisplay, "\n")},
						},
					},
				},
			},
		})

		if err != nil {
			h.interactionRespondWithErrorLogging(s, i.Interaction, "Something went wrong"+internalError)
			return
		}

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
							Name: "Event Management",
							Value: strings.Join([]string{
								"Event related utilities come under the events command, two types of events are currently supported",
								"`scoreboard-campaign` - where participants claim scores with screenshot proofs and ones with the highest accumulated score wins",
								"`tournament` - single elimination random matchup pvp tournament",
								"`/events list` - list active events, optionally pass argument to list all events",
								"`/events join` - join the event specified with the event ID, or join the only active event",
								"`/events bail` - leave an event specified with the event ID, or the only active event",
								"`/events list-participant` - list the participants of the event specified with the event ID, or the only active event",
								"mod only: `/events create` - creates an event, if start time is unspecified, current time will be used",
								"mod only: `/events activate` - activate an event by ID",
								"mod only: `/events deactivate` - deactivate an event by ID",
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
