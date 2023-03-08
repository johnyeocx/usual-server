package busdb

import (
	"database/sql"
	"time"

	"github.com/stripe/stripe-go/v74"
)

type PayoutDB struct {
	DB	*sql.DB
}

func (p *PayoutDB) InsertPayout(bId int, extAccountId int, sp stripe.Payout) (error) {
	query := `
	INSERT into business_payout 
	(
		amount, business_id, currency, status, arrival_date,
		stripe_payout_id, stripe_dest_id, type, external_account_id
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
	ON CONFLICT (stripe_payout_id) DO UPDATE 
	SET status=$4, arrival_date=$5, stripe_dest_id=$7, type=$8, external_account_id=$9
	`

	_, err := p.DB.Exec(
		query, 
		sp.Amount, 
		bId, 
		sp.Currency, 
		sp.Status, 
		time.Unix(sp.ArrivalDate, 0), 
		sp.ID, 
		sp.Destination.ID, 
		sp.Type,
		extAccountId,
	)

	return err
}