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


func (s *SubscriptionDB) CusOwnsSub(cusId int, subId int) (*models.Subscription, error) {

	query := `SELECT s.stripe_sub_id
	from customer as c
	JOIN subscription as s on c.customer_id=s.customer_id
	WHERE c.customer_id=$1 AND s.sub_id=$2`
	
	var sub models.Subscription
	err := s.DB.QueryRow(query, cusId, subId).Scan(&sub.StripeSubID)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *SubscriptionDB) InsertSubscriptions(subs *[]models.Subscription) (
	[]models.Subscription, error,
) {

	numCols := 5

	valueStrings := make([]string, 0, len(*subs))
    valueArgs := make([]interface{}, 0, len(*subs) * numCols)
	
    for i, sub := range (*subs) {
		j := i * numCols + 1
		valueString := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", j, j + 1, j + 2, j + 3, j + 4)
        valueStrings = append(valueStrings, valueString)
        valueArgs = append(valueArgs, sub.StripeSubID)
        valueArgs = append(valueArgs, sub.CustomerID)
        valueArgs = append(valueArgs, sub.PlanID)
        valueArgs = append(valueArgs, sub.StartDate)
        valueArgs = append(valueArgs, sub.CardID)
    }
	
    stmt := fmt.Sprintf(
		`INSERT into subscription (stripe_sub_id, customer_id, plan_id, start_date, card_id) VALUES %s RETURNING sub_id`, 
		strings.Join(valueStrings, ","))
	
	returnedSubs := []models.Subscription{}
    rows, err := s.DB.Query(stmt, valueArgs...)
	if err != nil {
		return nil, err
	}

	index := 0
	for rows.Next() {
		var subId int
		if err := rows.Scan(&subId); err != nil {
			// TO FIX
			continue
		}
		sub := (*subs)[index]
		sub.ID = subId
		returnedSubs = append(returnedSubs, sub)
	}

	return returnedSubs, nil
}

func (s *SubscriptionDB) DeleteSubscription(subId int) (error) {
	fmt.Println("Deleting sub id:", subId)
	stmt := `DELETE from subscription WHERE sub_id=$1`
	_, err := s.DB.Exec(stmt, subId)
	return err
}