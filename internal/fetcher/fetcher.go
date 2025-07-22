package fetcher

import (
	"alerts/config"
	"alerts/model"
	"alerts/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var Client *http.Client = &http.Client{
	Timeout: 10 * time.Second,
}
var ChatClient *http.Client = &http.Client{
	Timeout: 5 * time.Second,
}
var update_id int64 = -1

var countryNameMap = map[string]string{
	"gb":  "Great Britain 🇬🇧",
	"us":  "United States 🇺🇸",
	"in":  "India 🇮🇳",
	"ir":  "Iran 🇮🇷",
	"jp":  "Japan 🇯🇵",
	"all": "🌍 The World",
}

func FetchEarthQuake() *model.Data {
	
	URL := config.BotConf.USGSUrl
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		log.Println("request creation failed")
		panic("request creation failed")
		return nil
	}
	res, err := Client.Do(req)
	if err != nil {
		log.Println("error fetching the response")
		panic("error fetching the response")
		return nil
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("error reading the body")
		panic("error reading the body")
		return nil
	}
	data := new(model.Data)
	err = json.Unmarshal(body, data)
	if err != nil {
		log.Println("error unmarshalling the response")
		panic("error unmarshalling the response")
		return nil
	}
	return data
}


func FetchChatId() *model.ChatUsers {

	var URL = fmt.Sprintf("%s%s/getUpdates", config.BotConf.TelegramDomain, config.BotConf.BotToken)
	if update_id != -1 {
		URL = fmt.Sprintf("%s%s/getUpdates?offset=%d", config.BotConf.TelegramDomain, config.BotConf.BotToken, update_id+1)
	}
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		log.Println("request creation failed")
		panic("request creation failed")
		return nil
	}
	res, err := ChatClient.Do(req)
	if err != nil {
		log.Println("error fetching the response")
		return nil
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("error reading the body")
		panic("error reading the body")
		return nil
	}
	data := new(model.ChatUsers)
	err = json.Unmarshal(body, data)
	if err != nil {
		log.Println("error unmarshalling the response")
		panic("error unmarshalling the response")
		return nil
	}
	size := len(data.Results)
	for i := range size {
		if update_id < data.Results[i].UpdateId {
			update_id = data.Results[i].UpdateId
		}
		if data.Results[i].Msg != nil {
			user := new(model.InsertBotUser)
			user.UserName = data.Results[i].Msg.Chat.UserName
			user.ChatId = data.Results[i].Msg.Chat.Id
			if err := repository.InsertIntoTelegramBot(user); err != nil {
				log.Println("error inserting the database")
				panic(fmt.Sprintf("error inserting the record into database %s", err.Error()))
			}
		} else if data.Results[i].CallbackQuery != nil {
			if err = repository.UpdateCountryPreference(data.Results[i].CallbackQuery.Data, data.Results[i].CallbackQuery.From.Id); err != nil {
				log.Println("error updating the database")
				panic(fmt.Sprintf("error updating the country preference %s", err.Error()))
			} else {
				selectedCountrymessage := fmt.Sprintf("You will now get EarthQuake notification for: %s", countryNameMap[data.Results[i].CallbackQuery.Data])
				SendMessageToTelegram(data.Results[i].CallbackQuery.From.Id, selectedCountrymessage)
			}

		}

	}
	return data
}

func SendMessageToTelegram(chatId int64, message string) error {
	botToken := config.BotConf.BotToken
	telegramAPI := fmt.Sprintf("%s%s/sendMessage", config.BotConf.TelegramDomain, botToken)
	telegramMessage := &model.TelegramMessage{
		ChatID: chatId,
		Text:   message,
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
