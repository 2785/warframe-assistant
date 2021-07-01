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

	"github.com/2785/warframe-assistant/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
)

// registerCommandsCmd represents the registerCommands command
var registerCommandsCmd = &cobra.Command{
	Use:   "registerCommands",
	Short: "Register known commands with discord",
	RunE: func(cmd *cobra.Command, args []string) error {
		dg, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
		if err != nil {
			return err
		}

		discordEventHandler := &discord.EventHandler{}

		err = dg.Open()
		if err != nil {
			return err
		}

		err = discordEventHandler.RegisterInteractionCreateHandlers(dg)
		if err != nil {
			return err
		}

		for _, v := range discordEventHandler.Commands {
			_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", v)
			if err != nil {
				return err
			}
		}

		cmds, err := dg.ApplicationCommands(dg.State.User.ID, "")
		if err != nil {
			return err
		}

		spew.Dump(cmds)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCommandsCmd)
}
