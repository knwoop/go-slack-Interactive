package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/slack-go/slack"
)

type Env struct {
	// Port is port to listen. Default is 5000.
	Port string `envconfig:"PORT" default:"5000"`

	// BotToken is bot user token to access to slack API.
	SlackBotToken string `envconfig:"SLACK_BOT_TOKEN" required:"true"`

	// VerificationToken is used to validate interactive messages from slack.
	VerificationToken string `envconfig:"VERIFICATION_TOKEN" required:"true"`

	// BotID is bot user ID.
	BotID string `envconfig:"BOT_ID" required:"true"`

	// ChannelID is slack channel ID where bot is working.
	// Bot responses to the mention in this channel.
	ChannelID string `envconfig:"CHANNEL_ID" required:"true"`
}

func main() {
	os.Exit(run())
}

func run() int {
	var env Env
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return 1
	}
	api := slack.New(env.SlackBotToken)

	http.Handle("/interaction", interactionHandler{
		slackClient:       api,
		verificationToken: env.VerificationToken,
	})

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
