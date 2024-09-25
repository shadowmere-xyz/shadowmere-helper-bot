package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	ServiceURL      string
	ServiceUserName string
	ServicePassword string
	TelegramToken   string
)

type ErrorMessage struct {
	URL []string `json:"url"`
}

func main() {
	ServiceURL = os.Getenv("SERVICE_URL")
	ServiceUserName = os.Getenv("SERVICE_USERNAME")
	ServicePassword = os.Getenv("SERVICE_PASSWORD")
	TelegramToken = os.Getenv("TELEGRAM_TOKEN")

	if ServiceURL == "" || ServiceUserName == "" || ServicePassword == "" || TelegramToken == "" {
		log.Fatal("missing data from environment")
	}

	bot, err := tgbotapi.NewBotAPI(TelegramToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 5

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message.IsCommand() {
			continue
		}
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			servers := findServers(update.Message.Text)
			if len(servers) == 0 {
				servers = findServers(update.Message.Caption)
			}
			if len(servers) > 0 {
				reply(update, bot, fmt.Sprintf("I found %d servers in this message.", len(servers)))
				replyText := make([]string, len(servers))
				for i, server := range servers {
					err := addServer(server)
					if err != nil {
						replyText[i] = fmt.Sprintf("Error: %v\n", err)
					} else {
						replyText[i] = fmt.Sprintf("Added server %s\n", server)
					}
					log.Info(replyText[i])
				}
				for _, replyStr := range replyText {
					reply(update, bot, replyStr)
				}
			} else {
				if update.Message.Chat.Type == "private" {
					reply(update, bot, fmt.Sprintf("I could not find any servers in this message"))
				}
			}
		}
	}
}

func reply(update tgbotapi.Update, bot *tgbotapi.BotAPI, reply string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("error sending reply %v", err)
	}
}

func findServers(input string) []string {
	servers := []string{}
	r, err := regexp.Compile("(\n+\\s*|\\s+)(ss://[A-Za-z0-9]+=*@.+:\\d+|ss://[A-Za-z0-9]+)")
	if err != nil {
		log.Printf("error building RE %v", err)
		return nil
	}

	for _, address := range r.FindAllString(fmt.Sprintf("\n%s", input), -1) {
		servers = append(servers, strings.TrimSpace(address))
	}
	return servers
}

func addServer(server string) error {
	log.Infof("Adding server %s", server)
	client := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf("url=%s", server))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/proxies/", ServiceURL), data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "shadowmere-helper-bot")
	req.SetBasicAuth(ServiceUserName, ServicePassword)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		errors := ErrorMessage{}
		err := json.Unmarshal(body, &errors)
		if err != nil {
			return fmt.Errorf("request failed with code %d.\nError message is not in the expected format", resp.StatusCode)
		}

		return fmt.Errorf("could not add key %s because %s", server, errors.URL)
	}

	return nil
}
