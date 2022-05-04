package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	ServiceURL      string
	ServiceUserName string
	ServicePassword string
	TelegramToken   string
)

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
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			reply(update, bot, "working on it")

			servers := findServers(update.Message.Text)
			if len(servers) == 0 {
				servers = findServers(update.Message.Caption)
			}
			if len(servers) > 0 {
				reply_text := ""
				for _, server := range servers {
					err := addServer(server)
					if err != nil {
						reply_text += fmt.Sprintf("Error adding server [%s] with error: %v\n", server, err)
					} else {
						reply_text += fmt.Sprintf("Added server %s\n", server)
					}
				}
				reply(update, bot, reply_text)
			} else {
				reply(update, bot, fmt.Sprintf("I could not find any servers in this message"))
			}
		}
	}
}

func reply(update tgbotapi.Update, bot *tgbotapi.BotAPI, reply string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	msg.ReplyToMessageID = update.Message.MessageID

	bot.Send(msg)
}

func findServers(input string) []string {
	servers := []string{}
	r, err := regexp.Compile("ss://[A-Za-z\\d+/]+@.+:\\d+|ss://[A-Za-z\\d+/]+")
	if err != nil {
		log.Printf("error building RE %v", err)
		return nil
	}

	for _, address := range r.FindAllString(input, -1) {
		servers = append(servers, address)
	}
	return servers
}

func addServer(server string) error {
	client := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf("url=%s", server))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/proxies/", ServiceURL), data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(ServiceUserName, ServicePassword)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusBadRequest && bytes.Contains(body, []byte("This proxy was already imported")) {
		return fmt.Errorf("proxy already imported")
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	return nil
}
