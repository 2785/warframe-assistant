# warframe-assistant

A discord bot to host in game events

## Features

- Based on [bwmarrin/discordgo](https://github.com/bwmarrin/discordgo) & [spf13/cobra](https://github.com/spf13/cobra)
- Discord slash commands :)
- IGN management - associate the users in your discord server with their in game name
- Event management - host events, set start / end dates, allow your users to submit proofs to claim score, and moderators to verify / update / confirm the scores with emoji reactions, allow any user to check current event status
- Deployable to heroku, alternatively build the docker image and run in your preferred environment
- Use the `/help` command to learn more about the bot

## Supported event types
- Scoreboard Campaign
  - Users submit scores with proofs, mods to verify the scores, and leaderboard is generated with the sum of all submissions from a user
- Scoreboard Leaderboard (I know the name is whack)
  - Same with Scoreboard campaign except for only the top score from each user counts

## Configurations

Some configurations are taken from environment variables
- `BOT_TOKEN` - discord bot token, you may get one by creating your own discord bot, see [discord documentation](https://discord.com/developers/docs/intro) for more details`
- `DATABASE_URL` - database DSN to connect to your postgres database - the tables required can be created with the `./db.sql` script
- The bot requires the following discord scopes:
  - `bot`
  - `application.commands`
- and the following bot permissions:
  - `Send Messages`
  - `Manage Messages`
  - `Add Reactions`
  - `Embed Links`
- You may run the `registerCommands` command to register the slash commands with discord, slash commands are cached and these may take some time to get propagated to your servers as per [discord documentation](https://discord.com/developers/docs/interactions/slash-commands#registering-a-command), `BOT_TOKEN` will be required
## Caveats

- The db table names are unfortunately hard coded in the initialization steps in `./cmd/serveBot.go`, this may get taken out as configurable at a future date
- The emoji workflow in score verification may be replaced in the future when [this PR](https://github.com/bwmarrin/discordgo/pull/933) on [bwmarrin/discordgo](https://github.com/bwmarrin/discordgo) goes in
- If more event types were added, the command description in the `/events create` will have a problem due to exceeding word limit
- Function docs will be added one day (tm)
- Cache eviction is only done upon successful workflow - if someone initiated a verification dialog and never touched it again, the cache will not be invalidated, this can potentially lead to memory problems

## Next steps

This project is in active-ish development, my next steps are:
- Hook up redis proper, and / or add proper cache invalidation into the memory cache
- Put database initialization into the app
- Support PVP tournament

## Contribution

For feature requests / concerns feel free to make an issue, PR is welcome, if you think the documentation could be improved feel free to make an issue to request clarification as well.