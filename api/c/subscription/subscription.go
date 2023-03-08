package subscription

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	my_enums "github.com/johnyeocx/usual/server/constants/enums"
	"github.com/johnyeocx/usual/server/db"
	cusdb "github.com/johnyeocx/usual/server/db/cus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
	"github.com/stripe/stripe-go/v74"
)



func GetSubscriptionData(
	sqlDB *sql.DB,
	cusId int,
	productId int,
) (map[string]interface{}, *models.RequestError) {

	c := cusdb.CustomerDB{DB: sqlDB}
	total, err := c.GetTotalSpent(cusId, productId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 1. Get Invoices
	invoices, err := c.GetSubInvoices(cusId, productId, 20)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 2. Get Usages
	usages, err := c.GetSubscriptionUsages(cusId, productId, 20)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return map[string]interface{}{
		"total": total,
		"invoices": invoices,
		"usages": usages,
	}, nil
}


func CreateSubscription(
	sqlDB *sql.DB, 
	customerId int,
	cardId int,
	productId int, 
) (*models.CreateSubReturn, *models.RequestError) {
	
	// Get list of products + subplans
	s := db.SubscriptionDB{DB: sqlDB}
	c := cusdb.CustomerDB{DB: sqlDB}

	// make sure customer is not already subscribed
	err := c.CheckCusSubscribed(customerId, []int{productId})
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusConflict,
		}
	}

	subProduct, stripeBusId, err := s.GetCreateSubData(productId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// Get stripe customer id
	cusStripeId, cardStripeId, err := c.GetCustomerAndCardStripeId(customerId, cardId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	
	stripeSub, err := my_stripe.CreateSubscription(
		*cusStripeId, *stripeBusId, *cardStripeId, *subProduct,
	)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	stripeIn := stripeSub.LatestInvoice
	lastIn := models.Invoice{}
	now := time.Now()

	
	if stripeIn != nil {
		lastIn.CardID = cardId
		lastIn.Total = int(stripeIn.Total)
		lastIn.Created = time.Unix(stripeIn.Created, 0)
		lastIn.InvoiceURL = stripeIn.HostedInvoiceURL
		lastIn.Status = string(stripeIn.Status)
		lastIn.PaymentIntentStatus = my_enums.StripePMStatusToMYPMStatus(stripeSub.LatestInvoice.PaymentIntent.Status)
	}
	sub := models.Subscription{
		StripeSubID: stripeSub.ID,
		CustomerID: customerId,
		PlanID: subProduct.SubPlan.PlanID,
		StartDate: now,
		CardID: cardId,
		LastInvoice: &lastIn,
	}
	
	// // 4. INSERT INTO DB
	returnedSubs, err := s.InsertSubscriptions(&[]models.Subscription{sub})
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	sub.ID = returnedSubs[0].ID
	return &models.CreateSubReturn{
		Sub: sub,
		Status: stripeSub.LatestInvoice.PaymentIntent.Status,
		PaymentIntent: stripeSub.LatestInvoice.PaymentIntent,
	}, nil
}

func ResolvePaymentIntent(
	sqlDB *sql.DB,
	cusId int,
	cardId int,
	subId int,
) (map[string]interface{}, *models.RequestError) {

	// Get card
	c := cusdb.CustomerDB{DB: sqlDB}
	s := db.SubscriptionDB{DB: sqlDB}
	i := db.InvoiceDB{DB: sqlDB}

	card, err := c.GetCusCard(cusId, cardId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	sub, _, _, err := s.CusOwnsSub(cusId, subId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusForbidden,
		}
	}

	stripeSub, paymentIntent, err := my_stripe.UpdateSubDefaultCardAndConfirm(sub.StripeSubID, card.StripeID)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	err = s.UpdateSubCardID(subId, cardId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	err = i.UpdateInvoiceCardIDByStripeID(stripeSub.LatestInvoice.ID, cardId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
		
	}

	return map[string]interface{}{
		"payment_method_id": card.StripeID,
		"payment_intent": paymentIntent,
	}, nil
}

func GetPaymentIntent(
	sqlDB *sql.DB,
	cusId int,
	subId int,
) (map[string]interface{}, *models.RequestError) {

	// Get card
	s := db.SubscriptionDB{DB: sqlDB}

	_, _, lastIn, err := s.CusOwnsSub(cusId, subId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	paymentIntent, err := my_stripe.GetSubLastInvoicePaymentIntent(lastIn.PMIStripeID)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return map[string]interface{}{
		"payment_intent": paymentIntent,
	}, nil
}

func ResumeSubscription(
	sqlDB *sql.DB, 
	customerId int,
	cardId int,
	subId int,
) (*models.ResumeSubReturn, *models.RequestError) {


	c := cusdb.CustomerDB{DB: sqlDB}
	s := db.SubscriptionDB{DB: sqlDB}

	data, err := s.GetCusResumeSubData(customerId, subId)
	if err != nil {
		return  nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	
	plan := data["plan"].(models.SubscriptionPlan)
	sub := data["subscription"].(models.Subscription)

	if !sub.Cancelled {
		return nil, &models.RequestError{
			Err: errors.New("subscription is not cancelled so can't resume"),
			StatusCode: http.StatusBadRequest,
		}
	}


	var cardStripeId string

	if sub.CardID != cardId {
		card, err := c.GetCusCard(customerId, cardId)
		if err != nil {
			return nil, &models.RequestError{
				Err: err,
				StatusCode: http.StatusForbidden,
			}
		}
		cardStripeId = card.StripeID
	} else {
		cardStripeId = data["stripe_card_id"].(string)
	}


	cusStripeId := data["stripe_cus_id"].(string)

	busStripeId := data["stripe_bus_id"].(string)

	
	stripeSub, err := my_stripe.ResumeSubscription(
		cusStripeId,
		busStripeId,
		*plan.StripePriceID,
		cardStripeId,
		sub.Expires.Time,
	)


	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	err = s.ResumeSubscription(subId, cardId, stripeSub.ID)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	
	var paymentIntent *stripe.PaymentIntent
	var lastIn models.Invoice
	if stripeSub.LatestInvoice == nil {
		paymentIntent = nil
		
	} else {
		paymentIntent = stripeSub.LatestInvoice.PaymentIntent
		lastIn.Created = time.Unix(stripeSub.LatestInvoice.Created, 0)
		lastIn.CardID = cardId
		lastIn.Total = int(stripeSub.LatestInvoice.Total)
		lastIn.InvoiceURL = stripeSub.LatestInvoice.HostedInvoiceURL
		lastIn.Status = string(stripeSub.LatestInvoice.Status)
		lastIn.PaymentIntentStatus = my_enums.StripePMStatusToMYPMStatus(stripeSub.LatestInvoice.PaymentIntent.Status)
	}

	return &models.ResumeSubReturn{
		LastInvoice: &lastIn,
		PaymentIntent: paymentIntent,
	}, nil
}


func CancelSubscription(
	sqlDB *sql.DB,
	cusId int,
	subId int,
) (map[string]interface{}, *models.RequestError) {
	s := db.SubscriptionDB{DB: sqlDB}

	// 1. check if cus owns sub
	sub, plan, lastInvoice, err := s.CusOwnsSub(cusId, subId)
	if err != nil && err != sql.ErrNoRows {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	if err == sql.ErrNoRows {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadRequest,
		}
	}

	// 2. Delete if only one payment
	deleted, err := DeleteSubIfNoPayments(sqlDB, subId, sub.StripeSubID)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	if deleted {
		return map[string]interface{}{
			"deleted": true,
		}, nil
	}

	// 2. delete from stripe
	err = my_stripe.CancelSubscription(sub.StripeSubID)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 3. Set sql to cancelled and determine expired date
	recurring := plan.RecurringDuration
	var expires time.Time
	if (lastInvoice.PaymentIntentStatus != my_enums.PMIPaymentSucceeded) {
		expires = GetNextBillingDate(recurring, lastInvoice.Created)
	} else {
		expires = lastInvoice.Created
	}

	// // 3. update sql
	err = s.CancelSubscription(subId, expires)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return map[string]interface{}{
		"expires": expires,
	}, nil
}


func DeleteSubIfNoPayments(sqlDB *sql.DB, subId int, stripeSubId string) (bool, error) {
	s := db.SubscriptionDB{DB: sqlDB}

	invoices, err := s.GetSubInvoicesFromSubID(subId, 10)
	if err != nil {
		return false, err
	}

	
	if len(invoices) == 1 && invoices[0].PaymentIntentStatus != "succeeded" {
		err := s.DeleteSubscriptionAndInvoices(subId)

		if err != nil {
			return false, err
		}
		
		err = my_stripe.CancelSubscription(stripeSubId)
		return true, err
	}
	
	return false, nil
}

func ChangeSubDefaultCard(
	sqlDB *sql.DB,
	cusId int,
	subId int,
	cardId int,
) ( *models.RequestError) {
	s := db.SubscriptionDB{DB: sqlDB}
	c := cusdb.CustomerDB{DB: sqlDB}

	// 1. check if cus owns sub
	sub, _, _, err := s.CusOwnsSub(cusId, subId)
	if err != nil && err != sql.ErrNoRows {
		return  &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	if err == sql.ErrNoRows {
		return  &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadRequest,
		}
	}

	// 2. get card stripe id
	_, cardStripeId, err :=c.GetCustomerAndCardStripeId(cusId, cardId)
	if err != nil {
		return  &models.RequestError{
			Err: err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	// 2. change stripe default
	err = my_stripe.ChangeSubDefaultCard(sub.StripeSubID, *cardStripeId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// // 3. update sql
	err = s.UpdateSubCardID(subId, cardId)
	if err != nil {
		return  &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	return nil
}

func GetNextBillingDate(
	recurring models.TimeFrame, 
	lastInvoiceDate time.Time,

) (time.Time) {
	interval := recurring.Interval.String
	intervalCount := recurring.IntervalCount.Int16
	
	if recurring.Interval.String == "day" {
		return lastInvoiceDate.Add(time.Hour * 24 * time.Duration(intervalCount))
	} else if interval == "week" {
		return lastInvoiceDate.Add(time.Hour * 24 * 7 * time.Duration(intervalCount))
	} else if interval == "month" {
		// get next month
		next := lastInvoiceDate.AddDate(0, int(intervalCount), 0)
		if next.Month() - lastInvoiceDate.Month() > time.Month(intervalCount) {
			next = EndOfMonth(next.AddDate(0, -1, 0))
		}

		return next
	} else {
		return lastInvoiceDate.AddDate(1, 0, 0)
	}
}

func EndOfMonth(date time.Time) (time.Time) {
    return date.AddDate(0, 1, -date.Day())
}
