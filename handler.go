package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type interactionHandler struct {
	slackClient       *slack.Client
	verificationToken string
}

func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	verifier, err := slack.NewSecretsVerifier(r.Header, os.Getenv("SLACK_SIGNING_SECRET"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	if err := verifier.Ensure(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	body := buf.String()
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			_, _, _ = h.slackClient.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))

			message := strings.Split(ev.Text, " ")
			if len(message) < 2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			command := message[1]
			switch command {
			case "ping":
				if _, _, err := h.slackClient.PostMessage(ev.Channel, slack.MsgOptionText("pong", false)); err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
	default:
		log.Printf("[ERROR] ]Invalid event type was submitted: %s", eventsAPIEvent.Type)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
