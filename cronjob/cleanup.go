package cronjob

import (
	"alerts/repository"
	"log"

	"github.com/robfig/cron/v3"
)

func ScheduleCleanupJob() {
	c := cron.New()
	_, err := c.AddFunc("0 0 * * *", func() {
		err := repository.ClearAllUpdatesForADay()
		if err != nil {
			log.Println("Error cleaning up data:", err)
		} else {
			log.Println("Successfully cleaned up the data")
		}
	})
	if err != nil {
		log.Fatal("Failed to schedule job:", err)
	}
	c.Start()
}
