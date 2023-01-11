package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/johnyeocx/usual/server/db/models"
)

type SubscriptionDB struct {
	DB *sql.DB
}

func (s *SubscriptionDB) InsertSubscriptions(subs *[]models.Subscription) (error) {

	valueStrings := make([]string, 0, len(*subs))
    valueArgs := make([]interface{}, 0, len(*subs) * 4)
    for i, sub := range (*subs) {
		j := i * 4 + 1
		valueString := fmt.Sprintf("($%d, $%d, $%d, $%d)", j, j + 1, j + 2, j + 3)
        valueStrings = append(valueStrings, valueString)
        valueArgs = append(valueArgs, sub.StripeSubID)
        valueArgs = append(valueArgs, sub.CustomerID)
        valueArgs = append(valueArgs, sub.PlanID)
        valueArgs = append(valueArgs, sub.StartDate)
    }
	
    stmt := fmt.Sprintf(
		`INSERT into subscription (stripe_sub_id, customer_id, plan_id, start_date) VALUES %s`, 
		strings.Join(valueStrings, ","))

    _, err := s.DB.Exec(stmt, valueArgs...)

	if err != nil {
		return err
	}

	return nil
}