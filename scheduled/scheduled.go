package scheduled

import (
	"database/sql"
	"time"

	"github.com/go-co-op/gocron"
)

func RunCronJobs(db *sql.DB) {
	go DeleteExpiredOTPs(db)
}

func DeleteExpiredOTPs(db *sql.DB) {
	s := gocron.NewScheduler(time.UTC)
	s.Every(5).Minutes().Do(func() {
		deleteStatement := `
			DELETE FROM message_otps WHERE expiry > $1
		`

		db.Exec(deleteStatement, time.Now())
	})

	s.StartBlocking()
}