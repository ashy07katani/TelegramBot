package model

type Data struct {
	Features []*Feature `json:"features"`
}
type Feature struct {
	Id         string      `json:"id"`
	Geo        *Geometry   `json:"geometry"`
	Properties *Properties `json:"properties"`
}

type Geometry struct {
	Coordinates []float64 `json:"coordinates"`
}
type Properties struct {
	Title     string  `json:"title"`
	Magnitude float64 `json:"mag"`
	Place     string  `json:"place"`
	Tsunami   int     `json:"tsunami"`
	Time      int64   `json:"time"`
}

type ChatUsers struct {
	Results []*Result `json:"result"`
}

type Result struct {
	UpdateId      int64          `json:"update_id"`
	Msg           *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageID int64 `json:"message_id"`
	Chat      *Chat `json:"chat"`
}

type Chat struct {
	Id       int64  `json:"id"`
	UserName string `json:"username"`
}

type InsertAlertRequest struct {
	EarthQuakeId string
	ChatId       int64
}

type InsertBotUser struct {
	ChatId   int64
	UserName string
}

type TelegramMessage struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type GeoResponse struct {
	Address Address `json:"address"`
}

type Address struct {
	City        string `json:"city,omitempty"`
	County      string `json:"county,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code"`
}

type InlineKeyBoardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type InlineKeyBoardMarkup struct {
	InlineKeyBoard [][]InlineKeyBoardButton `json:"inline_keyboard"`
}

type TelegramMessageWithKeyboard struct {
	ChatID      int64                `json:"chat_id"`
	Text        string               `json:"text"`
	ReplyMarkup InlineKeyBoardMarkup `json:"reply_markup"`
}

type CallbackQuery struct {
	From *User  `json:"from"`
	Data string `json:"data"`
}

type User struct {
	Id int64 `json:"id"`
}
