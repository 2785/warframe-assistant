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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/2785/warframe-assistant/internal/discord"
	"github.com/2785/warframe-assistant/internal/scores"
	"github.com/2785/warframe-assistant/internal/stringcache"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// serveBotCmd represents the serveBot command
var serveBotCmd = &cobra.Command{
	Use:   "serveBot",
	Short: "TBD",
	Long:  `TBD`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dg, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
		if err != nil {
			return err
		}

		db, err := sqlx.Open("postgres", os.Getenv("DATABASE_URL"))
		if err != nil {
			return err
		}

		logger, err := zap.NewDevelopment(zap.Fields(zap.String("pl", "warframe-assistant"), zap.String("co", "serveBot")))
		if err != nil {
			return err
		}

		pgService := &scores.PostgresService{
			DB:     db,
			Table:  "devtest",
			Logger: logger,
		}
		memCache := stringcache.NewMemory(5*time.Minute, 10*time.Minute)
		discordEventHandler := &discord.EventHandler{
			RoleIDForVerification: "847537490033770497",
			Cache:                 memCache,
			Logger:                logger,
			Prefix:                "?!",
			EventScoreService:     pgService,
		}

		dg.Identify.Intents = discordgo.IntentsGuildMessages + discordgo.IntentsGuildMessageReactions

		dg.AddHandler(discordEventHandler.HandleMessageCreate)
		dg.AddHandler(discordEventHandler.HandleMessageReactionAdd)
		err = dg.Open()
		if err != nil {
			return err
		}

		logger.Info("server started")

		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc

		dg.Close()
		logger.Info("server terminated")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveBotCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveBotCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveBotCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
