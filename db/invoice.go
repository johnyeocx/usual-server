package db

import (
	"database/sql"

	"github.com/johnyeocx/usual/server/db/models"
)

type InvoiceDB struct {
	DB *sql.DB
}

func (i *InvoiceDB) GetSubFromStripeID (
	subStripeId string,
) (*models.Subscription, error) {

	var sub models.Subscription
	err := i.DB.QueryRow(`SELECT sub_id, card_id FROM subscription WHERE stripe_sub_id=$1`, subStripeId).Scan(
		&sub.ID, &sub.CardID)
	if err != nil {
		return nil, err
	}

	return &sub, nil
}

func (i *InvoiceDB) InsertInvoice (
	invoice *models.Invoice,
) (error) {

	_, err := i.DB.Exec(`INSERT into invoice 
		(
			stripe_in_id, stripe_cus_id, stripe_sub_id, stripe_price_id, stripe_prod_id,
			paid, status, attempted, total, created, invoice_url, app_fee_amt, default_payment_method, sub_id, card_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		invoice.InStripeID, invoice.CusStripeID, invoice.SubStripeID, invoice.PriceStripeID,
		invoice.ProdStripeID, invoice.Paid, invoice.Status, invoice.Attempted, invoice.Total, invoice.Created, 
		invoice.InvoiceURL, invoice.ApplicationFeeAmt, invoice.DefaultPaymentMethod, invoice.SubID, invoice.CardID,
	)

	return err
}