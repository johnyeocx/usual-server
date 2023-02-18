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

	query := `
		SELECT s.sub_id, s.customer_id, s.card_id, s.cancelled, p.name, b.name, sp.unit_amount
		FROM subscription as s 
		JOIN subscription_plan as sp on sp.plan_id=s.plan_id
		JOIN product as p on p.product_id=sp.product_id
		JOIN business as b on b.business_id=p.business_id
		WHERE stripe_sub_id=$1
	`
	var sub models.Subscription
	sub.SubProduct = &models.SubscriptionProduct{}
	sub.SubProduct.Product = models.Product{}
	sub.SubProduct.SubPlan = models.SubscriptionPlan{}
	
	err := i.DB.QueryRow(query, subStripeId).Scan(
		&sub.ID, 
		&sub.CustomerID,
		&sub.CardID,
		&sub.Cancelled,
		&sub.SubProduct.Product.Name,
		&sub.BusinessName,
		&sub.SubProduct.SubPlan.UnitAmount,
	)
	
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
			paid, status, attempted, total, created, invoice_url, 
			app_fee_amt, default_payment_method, sub_id, card_id, payment_intent_status, stripe_pmi_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) 
		ON CONFLICT (stripe_in_id) DO UPDATE 
		SET paid=$6, status = $7, payment_intent_status=$16
		` , 
		invoice.InStripeID, invoice.CusStripeID, invoice.SubStripeID, invoice.PriceStripeID, 
		invoice.ProdStripeID, invoice.Paid, invoice.Status, invoice.Attempted, invoice.Total, invoice.Created, 
		invoice.InvoiceURL, invoice.ApplicationFeeAmt, invoice.DefaultPaymentMethod, 
		invoice.SubID, invoice.CardID, invoice.PaymentIntentStatus, invoice.PMIStripeID,
	)

	return err
}

func (i *InvoiceDB) UpdateInvoiceCardIDByStripeID(inStripeID string, cardId int) (error) {
	stmt := `UPDATE invoice SET card_id=$1 WHERE stripe_in_id=$2`
	_, err := i.DB.Exec(stmt, cardId, inStripeID)
	return err
}

func (i *InvoiceDB) UpdateInvoiceStatus(inStripeID string, status string, paymentIntentStatus string) (error) {
	stmt := `UPDATE invoice SET status=$1, payment_intent_status=$2 WHERE stripe_in_id=$3`
	_, err := i.DB.Exec(stmt, status, paymentIntentStatus, inStripeID)
	return err
}