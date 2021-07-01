<a name="unreleased"></a>
## [Unreleased]


<a name="v0.3.1"></a>
## [v0.3.1] - 2021-07-01
### Bug Fixes
- **(deps):** update module github.com/spf13/cobra to v1.2.0

### Docs
- clean up CLI docs


<a name="v0.3.0"></a>
## [v0.3.0] - 2021-07-01
### Bug Fixes
- pull redis and database url from viper instead of envvar
- remove unused imports from go mod
- **(deps):** update module github.com/go-redis/redis/v8 to v8.11.0
- **(deps):** update module go.uber.org/zap to v1.18.1
- **(deps):** update module go.uber.org/zap to v1.18.0
- **(deps):** update module github.com/go-redis/redis/v8 to v8.10.0
- **(deps):** update github.com/hako/durafmt commit hash to 5c1018a

### Code Refactoring
- rework config, logger init

### Docs
- update readme

### Features
- verify workflow is now done with buttons


<a name="v0.2.0"></a>
## [v0.2.0] - 2021-06-13
### Bug Fixes
- empty role is now properly handled

### Docs
- add testing instructions to readme
- add changelog

### Features
- add redis caching


<a name="v0.1.0"></a>
## v0.1.0 - 2021-06-11
### Bug Fixes
- add in the leaderboard event type

### Docs
- add readme
- update the help command

### Features
- add the two scoreboard workflows into the interaction flow
- add a help message slash command
- add event, participation, ign crud, add slash commands
- add support for score events


[Unreleased]: https://github.com/2785/warframe-assistant/compare/v0.3.1...HEAD
[v0.3.1]: https://github.com/2785/warframe-assistant/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/2785/warframe-assistant/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/2785/warframe-assistant/compare/v0.1.0...v0.2.0
