package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Starshine113/crouter"
	"github.com/bwmarrin/discordgo"
)

var (
	token string
	owner string
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&owner, "o", "", "Bot Owner")
	flag.Parse()
}

func main() {
	if token == "" || owner == "" {
		log.Fatalln("One or more required flags was empty")
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("error creating Discord session,", err)
	}

	// create the router
	r := crouter.NewRouter(dg, []string{owner}, []string{"?", "!"})

	// add the message create handler
	dg.AddHandler(r.MessageCreate)

	// add a sample ping command
	r.AddCommand(&crouter.Command{
		Name: "Ping",

		Summary: "Check if the bot is running",

		Command: func(ctx *crouter.Ctx) (err error) {
			ping := ctx.Session.HeartbeatLatency().Round(time.Millisecond)
			_, err = ctx.Sendf("Ping! Average latency: %s", ping)
			return err
		},
	})

	// a sample panicking command
	r.AddCommand(&crouter.Command{
		Name: "Panic",

		Summary: "Make the bot panic",

		Command: func(ctx *crouter.Ctx) (err error) {
			panic("panicking!")
		},
	})

	// a sample command with arguments
	r.AddCommand(&crouter.Command{
		Name: "Arguments",

		Summary: "Play around with arguments",

		Command: func(ctx *crouter.Ctx) (err error) {
			_, err = ctx.Send(fmt.Sprintf("arg length => %v\narguments: %v\nraw arguments: %v", len(ctx.Args), ctx.Args, ctx.RawArgs))
			return err
		},
	})

	r.AddCommand(&crouter.Command{
		Name: "Permissions",

		Summary: "This command needs extra permissions",

		CustomPermissions: []func(*crouter.Ctx) (string, bool){
			func(ctx *crouter.Ctx) (string, bool) {
				if ctx.Author.ID == owner {
					return "", true
				}
				return "Bot Admin", false
			},
		},

		Command: func(ctx *crouter.Ctx) (err error) {
			_, err = ctx.Send("Congrats, you're a bot admin!")
			return err
		},
	})

	// this function is called after every succesful command execution
	r.PostFunc = postFunc

	// add intents
	dg.Identify.Intents = discordgo.MakeIntent(crouter.RequiredIntents)

	// open a connection to Discord
	err = dg.Open()
	if err != nil {
		panic(err)
	}
	defer dg.Close()

	log.Println("Connected to Discord. Press Ctrl-C or send an interrupt signal to stop.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
