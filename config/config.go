package config

type BotConfig struct {
	OpenstreetmapDomain string  `env:"openstreetmapDomain"`
	TelegramDomain      string  `env:"telegramDomain"`
	BotToken            string  `env:"botToken"`
	UserTicker          int     `env:"userTicker"`
	AlertTicker         int     `env:"alertTicker"`
	USGSUrl             string  `env:"usgsURL"`
	Magnitude           float64 `env:"magnitude"`
	MapURL              string  `env:"mapURL"`
}
type DBConfig struct {
	DatabaseName     string `env:"databaseName"`
	DriverName       string `env:"driverName"`
	PostgresPort     int    `env:"postgresPort"`
	PostgresUserName string `env:"postgresUserName"`
	PostgresPassword string `env:"postgresPassword"`
	PostgresHost     string `env:"postgresHost"`
}

var BotConf *BotConfig
var DBConf *DBConfig

/*

openstreetmapDomain=https://nominatim.openstreetmap.org/reverse
telegramDomain=https://api.telegram.org
botToken=7514104719:AAG11GRpGEzemBOUKq0lMKuujmMfuBrV-ms
databaseName=Earthquake
driverName=postgres
postgresPort=5432
postgresUserName=postgres
postgresPassword=root
postgresHost=localhost
userTicker=15
alertTicker=20

*/
