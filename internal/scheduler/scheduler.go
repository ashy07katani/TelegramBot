package scheduler

import (
	"alerts/config"
	"alerts/internal/fetcher"
	"alerts/model"
	"alerts/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var ticker = time.NewTicker(time.Second * 15)
var addUser = time.NewTicker(time.Second * 14)
var client http.Client
var addressClient http.Client

func PollAddUser(wg *sync.WaitGroup) {
	defer addUser.Stop()
	defer wg.Done()
	for {
		select {
		case <-addUser.C:
			log.Println("Fetching chatId from Telegram")
			fetcher.FetchChatId()
		}
	}
}

func PollingAlerts(wg *sync.WaitGroup) {
	defer ticker.Stop()
	defer wg.Done()
	for {
		select {
		case <-ticker.C:
			log.Println("Polling FetchEarthQuake function")
			data := fetcher.FetchEarthQuake()
			user := repository.GetFromTelegramBot()
			pollingAlertUtil(user, data)
		}
	}
}

func pollingAlertUtil(user []*model.InsertBotUser, data *model.Data) {

	size := len(user)
	dataSize := len(data.Features)
	addresses := []*model.Address{}

	for j := range dataSize {
		address, err := fetchLocation(data.Features[j].Geo)

		if err != nil {
			log.Println("Error fetching the location", err.Error())
			return
		}
		addresses = append(addresses, address)
		log.Println("Country code is ", addresses[j].CountryCode)
	}

	for i := range size {
		keyBoardSent, err := repository.GetKeyBoardSent(user[i].ChatId)
		if err != nil {
			log.Println("Error getting keyboard sent flag:", err)
			panic("can't get the flag value for keyboard sent")
		}
		country, err := repository.GetCountry(user[i].ChatId)
		fmt.Println("country fetched for ", user[i].UserName, " ", country)
		if err == nil && country == "" {
			if !keyBoardSent {
				if err = SendKeyBoard(user[i].ChatId); err != nil {
					log.Println("ERROR SENDING KEYBOARD TO TELEGRAM", err.Error())
					return
				} else {
					repository.SetKeyBoardSent(user[i].ChatId)
				}
			}

		} else if err != nil {
			log.Println("ERROR FETCHING THE PREFERRED COUNTRY", err.Error())
			return
		} else {
			for j := range dataSize {
				if country == addresses[j].CountryCode || country == "all" {

					count, err := repository.GetAlertCount(data.Features[j].Id, user[i].ChatId)

					if err != nil {
						log.Println("Not able to fetch count", err.Error())
						return
					} else if count == 0 {
						req := new(model.InsertAlertRequest)
						req.ChatId = user[i].ChatId
						req.EarthQuakeId = data.Features[j].Id
						err = repository.InsertIntoSentAlert(req)
						if err != nil {
							log.Println("Failed to insert alert data into database", err.Error())
							return
						}

						timeOccured := time.UnixMilli(data.Features[j].Properties.Time).UTC()
						formatted := fmt.Sprintf("Alert Time: %s", timeOccured.Format("2006-01-02 15:04:05 MST"))
						var tsunamiAlert string
						if data.Features[j].Properties.Tsunami == 0 {
							tsunamiAlert = "No"
						} else {
							tsunamiAlert = "Yes"
						}
						mapURL := fmt.Sprintf(config.BotConf.MapURL, data.Features[j].Geo.Coordinates[0], data.Features[j].Geo.Coordinates[1], data.Features[j].Geo.Coordinates[0], data.Features[j].Geo.Coordinates[1])

						message := fmt.Sprintf(
							`🌍 *Earthquake Alert\!* 🌍

*%s*

📍 *Location:* %s, %s, %s
📏 *Magnitude:* %s
🕒 *Time:* Alert Time: %s
📡 *Depth:* %s km
🌊🚨 *Tsunami Alert:* %s
🗺️ [Click here to view location](%s)

⚠️ *Stay Safe:*
\- Move to an open area away from buildings
\- Avoid elevators
\- Drop, Cover, and Hold On\!`,
							escapeMdV2(data.Features[j].Properties.Title),
							escapeMdV2(addresses[j].State),
							escapeMdV2(addresses[j].County),
							escapeMdV2(addresses[j].Country),
							escapeMdV2(fmt.Sprintf("%.2f", data.Features[j].Properties.Magnitude)),
							escapeMdV2(formatted),
							escapeMdV2(fmt.Sprintf("%.2f", data.Features[j].Geo.Coordinates[2])),
							escapeMdV2(tsunamiAlert),
							mapURL,
						)

						if err = SendAlertToTelegram(user[i].ChatId, message); err != nil {
							log.Println("ERROR SENDING MESSAGE TO TELEGRAM", err.Error())
							return
						}
					}

				}

			}
		}

	}
}

func SendAlertToTelegram(chatId int64, message string) error {
	botToken := config.BotConf.BotToken
	telegramAPI := fmt.Sprintf("%s%s/sendMessage", config.BotConf.TelegramDomain, botToken)

	telegramMessage := &model.TelegramMessage{
		ChatID:    chatId,
		Text:      message,
		ParseMode: "MarkdownV2",
	}

	body, err := json.Marshal(telegramMessage)
	if err != nil {
		log.Println("failed to marshal message:", err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, telegramAPI, bytes.NewBuffer(body))
	if err != nil {
		log.Println("error creating Telegram request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "earthquake-alert-bot/1.0")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("error sending request to Telegram:", err)
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading response from Telegram:", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram API returned status %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("telegram API error: %s", respBody)
	}

	log.Println("Telegram message sent successfully:", string(respBody))
	return nil
}

func fetchLocation(m *model.Geometry) (*model.Address, error) {

	openStreetMapDomain := config.BotConf.OpenstreetmapDomain
	url := fmt.Sprintf("%s?format=jsonv2&lat=%f&lon=%f", openStreetMapDomain, m.Coordinates[1], m.Coordinates[0])

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("User-Agent", "earthquake-alert/1.0")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}

	var address model.GeoResponse
	err = json.Unmarshal(body, &address)
	if err != nil {
		log.Println("Error unmarshalling response:", err)
		return nil, err
	}

	return &address.Address, nil
}

func SendKeyBoard(chatId int64) error {
	botToken := config.BotConf.BotToken

	telegramAPI := fmt.Sprintf("%s%s/sendMessage", config.BotConf.TelegramDomain, botToken)
	keyboard := model.InlineKeyBoardMarkup{
		InlineKeyBoard: [][]model.InlineKeyBoardButton{
			{
				{Text: "🇬🇧 UK", CallbackData: "gb"},
				{Text: "🇺🇸 USA", CallbackData: "us"},
			},
			{
				{Text: "🇮🇳 India", CallbackData: "in"},
				{Text: "🇮🇷 Iran", CallbackData: "ir"},
			},
			{
				{Text: "🇯🇵 Japan", CallbackData: "jp"},
				{Text: "🇮🇹 Italy", CallbackData: "it"},
			},
			{
				{Text: "🇮🇩 Indonesia", CallbackData: "id"},
				{Text: "🇷🇺 Russia", CallbackData: "ru"},
			}, {
				{Text: "🌍 Global", CallbackData: "all"},
			},
		},
	}
	msg := model.TelegramMessageWithKeyboard{
		ChatID:      chatId,
		Text:        "Please select your preferred country for earthquake alerts:",
		ReplyMarkup: keyboard,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		log.Println("failed to marshal message:", err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, telegramAPI, bytes.NewBuffer(body))
	if err != nil {
		log.Println("error creating Telegram request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "earthquake-alert-bot/1.0")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("error sending request to Telegram:", err)
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading response from Telegram:", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram API returned status %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("telegram API error: %s", respBody)
	}

	log.Println("Telegram message sent successfully:", string(respBody))
	return nil
}

func escapeMdV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
