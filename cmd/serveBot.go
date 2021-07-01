/*
Copyright © 2021 Shiqi Zhao <zhao.shiqi.art@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/2785/warframe-assistant/internal/cache"
	"github.com/2785/warframe-assistant/internal/discord"
	"github.com/2785/warframe-assistant/internal/meta"
	"github.com/2785/warframe-assistant/internal/scores"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	databaseURL string
	redisURL    string
	botToken    string
	logLevel    string
)

// serveBotCmd represents the serveBot command
var serveBotCmd = &cobra.Command{
	Use:   "serveBot",
	Short: "TBD",
	Long:  `TBD`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zapConf := zap.NewProductionConfig()
		l := zapcore.InfoLevel

		err := l.Set(viper.GetString("log_level"))
		if err != nil {
			return err
		}
		zapConf.Level.SetLevel(l)

		logger, err := zapConf.Build(
			zap.Fields(zap.String("pl", "warframe-assistant"), zap.String("co", "serveBot")),
		)
		if err != nil {
			return err
		}

		if !viper.IsSet("bot_token") {
			return errors.New("Bot token must be supplied")
		}
		dg, err := discordgo.New("Bot " + viper.GetString("bot_token"))
		if err != nil {
			return err
		}

		if !viper.IsSet("database_url") {
			return errors.New("Database URL must be supplied")
		}
		db, err := sqlx.Open("postgres", viper.GetString("database_url"))
		if err != nil {
			return err
		}
		logger.Info("connected to postgres")

		var c cache.Cache

		if viper.IsSet("redis_url") {
			c, err = cache.NewRedis(viper.GetString("redis_url"), 10*time.Minute)
			if err != nil {
				return err
			}
			logger.Info("connected to redis")
		} else {
			c = cache.NewMemory(10 * time.Minute)
			logger.Warn("redis URL not found, using in memory cache")
		}

		pgService := &scores.PostgresService{
			DB:                     db,
			Logger:                 logger,
			ScoresTableName:        "event_scores",
			ParticipationTableName: "participation",
			UserIGNTableName:       "users",
		}

		metadataService := meta.NewWithCache(
			&meta.PostgresService{
				DB:                 db,
				ActionRoleTable:    "role_lookup",
				IGNTable:           "users",
				EventsTable:        "events",
				ParticipationTable: "participation",
				Logger:             logger.With(zap.String("co", "metadata-service-pg"))},
			cache.Named("meta", c),
			logger.With(zap.String("co", "metadata-service-cache")))

		discordEventHandler := &discord.EventHandler{
			Cache:             cache.Named("dialog", c),
			Logger:            logger,
			Prefix:            "?!",
			EventScoreService: pgService,
			MetadataService:   metadataService,
		}

		dg.Identify.Intents =
			discordgo.IntentsGuildMessages +
				discordgo.IntentsGuildIntegrations

		dg.AddHandler(discordEventHandler.HandleMessageCreate)
		dg.AddHandler(discordEventHandler.HandleInteractionsCreate)
		err = dg.Open()
		if err != nil {
			return err
		}

		err = discordEventHandler.RegisterInteractionCreateHandlers(dg)

		if err != nil {
			return err
		}

		logger.Info(
			"established websocket to discord",
			zap.String("session-id", dg.State.SessionID),
		)

		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc

		dg.Close()
		logger.Info("server terminated")
		return nil
	},
}

func init() {
	serveBotCmd.Flags().
		StringVar(&databaseURL, "database-url", "", "Database URL to connect to a Postgres instance")
	err := viper.BindPFlag("database_url", serveBotCmd.Flags().Lookup("database-url"))
	if err != nil {
		panic(err)
	}

	serveBotCmd.Flags().StringVar(&redisURL, "redis-url", "", "URL to connect to redis")
	err = viper.BindPFlag("redis_url", serveBotCmd.Flags().Lookup("redis-url"))
	if err != nil {
		panic(err)
	}

	serveBotCmd.Flags().StringVar(&botToken, "bot-token", "", "URL to connect to redis")
	err = viper.BindPFlag("bot_token", serveBotCmd.Flags().Lookup("bot-token"))
	if err != nil {
		panic(err)
	}

	serveBotCmd.Flags().
		StringVar(&logLevel, "log-level", "info", "Log level, select between:\n"+strings.Join(
			funk.Map(
				[]zapcore.Level{
					zapcore.DebugLevel,
					zapcore.InfoLevel,
					zapcore.WarnLevel,
					zapcore.ErrorLevel,
					zapcore.DPanicLevel,
					zapcore.PanicLevel,
					zapcore.FatalLevel,
				}, func(l zapcore.Level) string { return l.String() }).([]string),
			"\n",
		))
	err = viper.BindPFlag("log_level", serveBotCmd.Flags().Lookup("log-level"))
	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(serveBotCmd)
}
