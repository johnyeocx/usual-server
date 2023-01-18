package db

import (
	"database/sql"

	"github.com/johnyeocx/usual/server/db/models"
)

type InvoiceDB struct {
	DB *sql.DB
}

func (i *InvoiceDB) InsertInvoice (
	invoice *models.Invoice,
) (error) {

	_, err := i.DB.Exec(`INSERT into invoice 
		(
			stripe_in_id, stripe_cus_id, stripe_sub_id, stripe_price_id, stripe_prod_id,
			paid, status, attempted, total, created, invoice_url, app_fee_amt
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		invoice.InStripeID, invoice.CusStripeID, invoice.SubStripeID, invoice.PriceStripeID,
		invoice.ProdStripeID, invoice.Paid, invoice.Status, invoice.Attempted, invoice.Total, invoice.Created, 
		invoice.InvoiceURL, invoice.ApplicationFeeAmt,
	)

	return err
}