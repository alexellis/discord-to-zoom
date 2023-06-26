package function

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	sdk "github.com/openfaas/go-sdk"
)

var usernames []string

var CHANNEL_MESSAGE_WITH_SOURCE = float64(4)

func init() {

	if v, ok := os.LookupEnv("discord_usernames"); ok && len(v) > 0 {
		parts := strings.Split(v, ",")
		for _, u := range parts {
			usernames = append(usernames, strings.TrimSpace(u))
		}
	}

	log.Println("Authorized users", usernames)

	discordClientID := os.Getenv("discord_client_id")
	if len(discordClientID) == 0 {
		panic("discord_client_id not set")
	}

	registerCommand := fmt.Sprintf("https://discord.com/api/v10/applications/%s/commands",
		discordClientID)

	botToken, err := sdk.ReadSecret("discord-bot-token")
	if err != nil {
		panic(err)
	}

	commandOptions := `{"name": "zoom", "description": "Create a Zoom meeting", "options": [{"name": "topic", "description": "The topic of the meeting", "type": 3, "required": false}]}`

	req, err := http.NewRequest(http.MethodPost, registerCommand, strings.NewReader(commandOptions))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "DiscordBot (https://alexellis.io, 1)")
	req.Header.Set("Authorization", "Bot "+botToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	log.Printf("Command status: %d", res.StatusCode)

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		log.Println("Unable to register command", res.Status)
		return
	}

	log.Println("Registered command")
}

// Handle receives a HTTP request from a Discord command, validates it
// and if authorized, creates a Zoom meeting, returning the URL and
// the unique passcode to the users in the channel.
func Handle(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		defer r.Body.Close()
		body, _ = io.ReadAll(r.Body)
	}

	discordMsg := make(map[string]interface{})

	if err := json.Unmarshal(body, &discordMsg); err != nil {
		log.Println("Unable to unmarshal Discord message ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if v, ok := discordMsg["type"].(float64); ok && v == 1 {
		verify(w, r, body)
		return
	}

	if os.Getenv("print_input") == "true" {
		log.Println("Input\n", body)
	}

	msg := DiscordInteraction{}

	if err := json.Unmarshal(body, &msg); err != nil {
		log.Println("Unable to unmarshal Discord message ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	command := msg.Data.Name

	if command != "zoom" {
		http.Error(w, "invalid command", http.StatusBadRequest)
		return
	}

	if !validUser(msg) {
		log.Printf("User %s is not authorized to create a Zoom meeting", msg.Member.User.Username)

		commandRes := DiscordResponse{
			Type: CHANNEL_MESSAGE_WITH_SOURCE,
			Data: DiscordResponseData{
				Content: "Sorry, you're not authorized to create a Zoom meeting right now.",
			},
		}
		writeResponse(w, commandRes)
		return
	}

	commandValue := ""
	for _, p := range msg.Data.Options {
		if p.Name == "topic" {
			if v, ok := p.Value.(string); ok {
				commandValue = v
			}
			break
		}
	}

	z, err := createMeeting(commandValue)
	if err != nil {
		commandRes := DiscordResponse{
			Type: CHANNEL_MESSAGE_WITH_SOURCE,
			Data: DiscordResponseData{
				Content: fmt.Sprintf("Sorry, we couldn't create the Zoom meeting, error: %s", err.Error()),
			},
		}

		writeResponse(w, commandRes)
		return
	}

	commandRes := DiscordResponse{
		Type: CHANNEL_MESSAGE_WITH_SOURCE,
		Data: DiscordResponseData{
			Content: fmt.Sprintf("A Zoom meeting has been created\n\nTopic: %s\nMeeting ID: %d\nPassword: %s\n",
				z.Topic,
				z.ID,
				z.Password),
			Embeds: []DiscordResponseEmbed{
				{
					Title:       z.Topic,
					Description: "Join the Zoom call",
					URL:         z.JoinURL,
					Type:        "link",
				},
			},
		},
	}

	log.Printf("Created a Zoom call for: %s", z.Topic)
	writeResponse(w, commandRes)
}

func validUser(msg DiscordInteraction) bool {

	username := msg.Member.User.Username

	for _, u := range usernames {
		if u == username {
			return true
		}
	}
	return false
}

// writeResponse is a convenience function to writeResponse a response back to Discord
// in JSON format
func writeResponse(w http.ResponseWriter, commandRes DiscordResponse) {
	data, _ := json.Marshal(commandRes)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
