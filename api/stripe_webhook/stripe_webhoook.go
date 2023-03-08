package stripe_webhook

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"github.com/johnyeocx/usual/server/api/c/subscription"
	my_enums "github.com/johnyeocx/usual/server/constants/enums"
	"github.com/johnyeocx/usual/server/db"
	busdb "github.com/johnyeocx/usual/server/db/bus_db"
	cusdb "github.com/johnyeocx/usual/server/db/cus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/fcm"
	"github.com/stripe/stripe-go/v74"
)

func VoidedInvoice(sqlDB *sql.DB, fbApp *firebase.App, data map[string]interface{}) (error) {
	invoice := ParseInvoicePaid(data)
	i := db.InvoiceDB{DB: sqlDB}
	c := cusdb.CustomerDB{DB: sqlDB}

	if !invoice.SubStripeID.Valid {
		return errors.New("no subscription stripe id")
	}

	sub, err := i.GetSubFromStripeID(invoice.SubStripeID.String)
	if err != nil {
		return err
	}

	invoice.PaymentIntentStatus = my_enums.PMIPaymentCancelled
	err = i.InsertInvoice(invoice)
	if err != nil {
		return err
	}
	
	if (sub.Cancelled) {
		return nil
	}
	
	_, reqErr := subscription.CancelSubscription(sqlDB, sub.CustomerID, sub.ID)
	if reqErr != nil {
		log.Println("Failed to cancel subscription")
		return reqErr.Err
	}


	// SEND PUSH NOTIFICATION
	fcmToken, err := c.GetCusFCMToken(sub.CustomerID)
		if err == sql.ErrNoRows {
			// handle no fcm token
		} else if err != nil {
			// do something else
			return  err
		} else {
			fcm.SendPaymentVoidedNotification(fbApp, *fcmToken, sub.ID, sub.SubProduct.Product.Name, *sub.BusinessName)
		}

	return nil
}

func InsertInvoice(
	sqlDB *sql.DB, 
	fbApp *firebase.App,
	data map[string]interface{}, 
	paymentStatus my_enums.MyPaymentIntentStatus,
) (*models.Invoice, error) {
	invoice := ParseInvoicePaid(data)
	invoice.PaymentIntentStatus = paymentStatus
	i := db.InvoiceDB{DB: sqlDB}
	c := cusdb.CustomerDB{DB: sqlDB}

	if (invoice.SubStripeID.Valid) {
		sub, err := i.GetSubFromStripeID(invoice.SubStripeID.String)
		if err != nil {
			return nil, err
		}
		invoice.SubID = sub.ID
		invoice.CardID = sub.CardID
		

		// SEND PUSH NOTIFICATION
		fcmToken, err := c.GetCusFCMToken(sub.CustomerID)
		if err == sql.ErrNoRows {
			// handle no fcm token
		} else if err != nil {
			// do something else
			return nil, err
		} else {
			if paymentStatus == my_enums.PMIPaymentFailed {
				fcm.SendPaymentFailedNotification(fbApp, *fcmToken, sub.ID, sub.SubProduct.Product.Name, *sub.BusinessName)
			} else if paymentStatus == my_enums.PMIPaymentSucceeded {
				fcm.SendPaymentSucceededNotification(fbApp, *fcmToken, sub.ID, sub.SubProduct.SubPlan.UnitAmount, sub.SubProduct.Product.Name, *sub.BusinessName)
			}
		}
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

func SetIndVerificationDocRequired(sqlDB *sql.DB, account stripe.Account, required bool) (*models.RequestError) {
	busStripeID := account.ID
	b := busdb.BusinessDB{DB: sqlDB}
	i, err := b.GetIndividualFromStripeID(busStripeID)

	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	err = b.UpdateIndividualVerificationDocumentRequired(i.ID, true)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return nil
}

func VerificationDocRequired(requirements []string) (bool){
	for _, r := range requirements {
		if strings.Contains(r, "verification.document") {
			return true
		}
	}

	return false
}