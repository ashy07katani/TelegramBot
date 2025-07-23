package repository

import (
	"alerts/config"
	"alerts/model"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

const (
	driverName = "postgres"
	host       = "localhost"
	port       = 5432
	user       = "postgres"
	password   = "root"
	dbname     = "Earthquake"
)

func DBInit(config *config.DBConfig) (*sql.DB, error) {

	driverSourceName := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.PostgresHost, config.PostgresPort, config.PostgresUserName, config.PostgresPassword, config.DatabaseName, config.SSLMode)
	DB, _ = sql.Open(config.DriverName, driverSourceName)
	err := DB.Ping()
	if err == nil {
		log.Println("Connection established successfully")
	}
	return DB, err
}

func InsertIntoTelegramBot(user *model.InsertBotUser) error {
	query := `insert into telegramuser (id, username) values ($1, $2) on conflict (id) do nothing`
	if DB == nil {
		log.Println("DB is nil what the heck")
		return errors.New("DB is nil")
	}
	_, err := DB.Exec(query, user.ChatId, user.UserName)
	if err != nil {
		log.Println("error inserting the data", err.Error())
		return err
	}
	return nil
}

func UpdateCountryPreference(country string, id int64) error {
	var err error
	query := `update telegramuser set country = $1 where id = $2`
	_, err = DB.Exec(query, country, id)
	if err != nil {
		log.Println("error updating prefered country", err.Error())
	}
	return err
}

func GetFromTelegramBot() []*model.InsertBotUser {
	botUsers := []*model.InsertBotUser{}
	query := "select id, username from telegramuser"
	row, err := DB.Query(query)
	if err != nil {
		log.Println("error fetching bot users from db: ", err.Error())
		return botUsers
	}
	defer row.Close()
	for row.Next() {
		var id int64
		var username sql.NullString
		user := new(model.InsertBotUser)

		if err = row.Scan(&id, &username); err != nil {
			log.Println("error populating the value into variables: ", err.Error())
			continue // Skip this iteration
		}

		user.ChatId = id
		if username.Valid {
			user.UserName = username.String
		} else {
			user.UserName = "" // or set a fallback like "unknown"
		}

		botUsers = append(botUsers, user)
	}

	return botUsers
}

func InsertIntoSentAlert(req *model.InsertAlertRequest) error {
	query := `insert into sent_alerts (earthquake_id, chat_id) values ($1, $2) on conflict (earthquake_id, chat_id) do nothing`
	_, err := DB.Exec(query, req.EarthQuakeId, req.ChatId)
	if err != nil {
		log.Println("error inserting the data", err.Error())
		return err
	}
	return nil
}
func GetAlertCount(quakeId string, chatId int64) (int, error) {
	var count int
	countQuery := `select count(*) from sent_alerts where earthquake_id = $1 and chat_id = $2`
	err := DB.QueryRow(countQuery, quakeId, chatId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, err
}

func GetCountry(chatId int64) (string, error) {
	var err error
	var countryCode sql.NullString
	countryQuery := `select country from telegramuser where id=$1`
	err = DB.QueryRow(countryQuery, chatId).Scan(&countryCode)
	if countryCode.Valid {
		return countryCode.String, err
	} else {
		return "", err
	}

}

func GetKeyBoardSent(chatId int64) (bool, error) {
	var err error
	var isKeyBoardSent bool
	getKeyBoardSentQuery := `select keyboardsent from telegramuser where id=$1;`
	err = DB.QueryRow(getKeyBoardSentQuery, chatId).Scan(&isKeyBoardSent)
	return isKeyBoardSent, err
}
func SetKeyBoardSent(chatId int64) error {
	query := `UPDATE telegramuser SET keyboardsent = true WHERE id = $1;`
	_, err := DB.Exec(query, chatId)
	if err != nil {
		log.Println("Error setting keyboard sent flag:", err)
	}
	return err
}

func ClearAllUpdatesForADay() error {
	query := `DELETE FROM sent_alerts WHERE inserted_at < NOW() - INTERVAL '1 day'`
	_, err := DB.Exec(query)
	if err != nil {
		log.Println("Error clearing notification for last 24 hours", err)
	}
	return err
}
