# Warframe Assistant

A discord bot to host in game events

## Features

- Based on [bwmarrin/discordgo](https://github.com/bwmarrin/discordgo) & [spf13/cobra](https://github.com/spf13/cobra)
- Discord slash commands :)
- BUTTONS
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

Configurations can be taken from either environment variable, command line arguments, or a config file, they need to be in all caps if they were to be taken from environment variables. Otherwise a wide variety of file formats can be used by starting the server with, for example, `... serveBot --config ./conf.json`. Config file defaults to `$HOME/.warframe-assistant.yaml`.

Configs:

|      name      | description                                                                                                                                                  | required | default |
| :------------: | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | :------: | ------- |
|  `bot_token`   | Discord bot token, you may get one by creating your own discord bot, see [discord documentation](https://discord.com/developers/docs/intro) for more details |   yes    |         |
| `database_url` | Database DSN to connect to your Postgres database, the required tables can be created with the `/db.sql` script                                              |   yes    |         |
|  `redis_url`   | DSN to connect to a redis instance, if omitted, an in memory cache will be used                                                                              |    no    |         |
|  `log_level`   | Log level of the zap logger used, see [here](https://pkg.go.dev/go.uber.org/zap/zapcore#Level) for a list of available levels                                |    no    | `info`  |

If you are hosting the bot yourself, you will need the following scopes and bot permissions to add it to a server:

- Scopes:
  - `bot`
  - `application.commands`
- Bot permissions:
  - `Send Messages`
  - `Manage Messages`
  - `Add Reactions`
  - `Embed Links`
- You may run the `registerCommands` command to register the slash commands with discord, slash commands are cached and these may take some time to get propagated to your servers as per [discord documentation](https://discord.com/developers/docs/interactions/slash-commands#registering-a-command), `bot_token` will be required

## Caveats

- The db table names are unfortunately hard coded in the initialization steps in `./cmd/serveBot.go`, this may get taken out as configurable at a future date
- If more event types were added, the command description in the `/events create` will have a problem due to exceeding word limit
- Function docs will be added one day (tm)

## Contribution

For feature requests / concerns feel free to make an issue, PR is welcome, if you think the documentation could be improved feel free to make an issue to request clarification as well. I'm looking for people to collaborate on the project.

## Testing

Running tests requires access to docker - [ory/dockertest](https://github.com/ory/dockertest) is used and it will automatically detect docker access most of the times, if test crashed it might leave hanging docker containers running postgres, you need to purge those manually if that happened.
