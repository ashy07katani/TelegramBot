package main

import (
	"alerts/config"
	"alerts/cronjob"
	"alerts/internal/scheduler"
	"alerts/repository"
	"log"
	"sync"

	env "github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

func main() {
	var wg sync.WaitGroup
	err := godotenv.Load("local.env")
	if err != nil {
		log.Println("Failed to load ENV file")
	}
	config.BotConf = &config.BotConfig{}
	config.DBConf = &config.DBConfig{}
	env.Parse(config.BotConf)
	env.Parse(config.DBConf)
	_, err = repository.DBInit(config.DBConf)
	if err != nil {
		log.Println("Cannot establish connection", err.Error())
	}
	cronjob.ScheduleCleanupJob()
	wg.Add(1)
	go scheduler.PollingAlerts(&wg)
	wg.Add(1)
	go scheduler.PollAddUser(&wg)
	wg.Wait()
}
