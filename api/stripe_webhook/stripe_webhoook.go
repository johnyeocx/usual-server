package stripe_webhook

import (
	"database/sql"
	"time"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
)

func InsertInvoice(sqlDB *sql.DB, data map[string]interface{}, paymentStatus string) (*models.Invoice, error) {
	invoice := ParseInvoicePaid(data)
	
	invoice.PaymentIntentStatus = paymentStatus
	i := db.InvoiceDB{DB: sqlDB}

	if (invoice.SubStripeID.Valid) {
		sub, err := i.GetSubFromStripeID(invoice.SubStripeID.String)
		if err != nil {
			return nil, err
		}
		invoice.SubID = sub.ID
		invoice.CardID = sub.CardID
	}
	
	
	err := i.InsertInvoice(invoice)
	return invoice, err
}

func ParseInvoicePaid(data map[string]interface{})(*models.Invoice) {

	var subStripeId models.JsonNullString
	if data["subscription"] == nil {
		subStripeId.Valid = false;
	} else {
		subStripeId.String = data["subscription"].(string)
		subStripeId.Valid = true
	}

	products := data["lines"].(map[string]interface{})["data"].([]interface{})
	product := products[0].(map[string]interface{})
	priceStripeId := product["price"].(map[string]interface{})["id"]
	prodStripeId := product["price"].(map[string]interface{})["product"]

	total := int(data["total"].(float64))
	createdUnix := int(data["created"].(float64))
	createdTimestamp := time.Unix(int64(createdUnix), 0)

	var appFeeAmt models.JsonNullInt64
	if data["application_fee_amount"] == nil {
		appFeeAmt.Valid = false;
	} else {
		appFeeAmt.Int64 = data["subscription"].(int64)
		appFeeAmt.Valid = true
	}


	var defaultPM models.JsonNullString
	if data["default_payment_method"] == nil {
		defaultPM.Valid = false;
	} else {
		defaultPM.String = data["default_payment_method"].(string)
		defaultPM.Valid = true
	}
	
	invoice := models.Invoice{
		InStripeID: data["id"].(string),
		CusStripeID: data["customer"].(string),
		SubStripeID: subStripeId,
		PMIStripeID: data["payment_intent"].(string),
		PriceStripeID: priceStripeId.(string),
		ProdStripeID: prodStripeId.(string),
		Paid: data["paid"].(bool),
		Status: data["status"].(string),
		Attempted: data["attempted"].(bool),
		Total:	total,
		Created: createdTimestamp,
		InvoiceURL: data["hosted_invoice_url"].(string),
		ApplicationFeeAmt: appFeeAmt,
		DefaultPaymentMethod: defaultPM,
	}

	return &invoice
}
