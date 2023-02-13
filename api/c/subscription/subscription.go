package subscription

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
)


func GetSubscriptionData(
	sqlDB *sql.DB,
	cusId int,
	productId int,
) (map[string]interface{}, *models.RequestError) {

	c := db.CustomerDB{DB: sqlDB}
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
	productIds []int, 
) ([]models.Subscription, *models.RequestError) {

	// Get list of products + subplans
	b := db.BusinessDB{DB: sqlDB}
	c := db.CustomerDB{DB: sqlDB}

	// make sure customer is not already subscribed
	err := c.CheckCusSubscribed(customerId, productIds)
	
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusConflict,
		}
	}

	subProducts, err := b.GetSubProductsFromIds(productIds)
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

	subscriptions := []models.Subscription{}
	now := time.Now()
	// today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// 3. INSERT INTO STRIPE
	for i := 0; i < len(*subProducts); i++ {
		stripeBusinessId, err := b.GetBusinessStripeID((*subProducts)[i].Product.BusinessID)
		if err != nil {
			return nil, &models.RequestError{
				Err: err,
				StatusCode: http.StatusBadGateway,
			}
		}	

		stripeSubId, err := my_stripe.CreateSubscription(
			*cusStripeId, *stripeBusinessId, *cardStripeId, (*subProducts)[i],
		)
		if err != nil {
			return nil, &models.RequestError{
				Err: err,
				StatusCode: http.StatusBadGateway,
			}
		}
		subscription := models.Subscription{
			StripeSubID: *stripeSubId,
			CustomerID: customerId,
			PlanID: (*subProducts)[i].SubPlan.PlanID,
			StartDate: now,
			CardID: cardId,
		}
		subscriptions = append(subscriptions, subscription)
	}
	
	// 4. INSERT INTO DB
	s := db.SubscriptionDB{DB: sqlDB}
	returnedSubs, err := s.InsertSubscriptions(&subscriptions)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	return returnedSubs, nil
}

func ResumeSubscription(
	sqlDB *sql.DB, 
	customerId int,
	cardId int,
	subId int,
) ( *models.RequestError) {


	c := db.CustomerDB{DB: sqlDB}
	s := db.SubscriptionDB{DB: sqlDB}

	data, err := s.GetCusResumeSubData(customerId, subId)
	if err != nil {
		return  &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	plan := data["plan"].(models.SubscriptionPlan)
	sub := data["subscription"].(models.Subscription)

	if !sub.Cancelled {
		return &models.RequestError{
			Err: errors.New("subscription is not cancelled so can't resume"),
			StatusCode: http.StatusBadRequest,
		}
	}


	var cardStripeId string

	if sub.CardID != cardId {
		card, err := c.GetCusCard(customerId, cardId)
		if err != nil {
			return &models.RequestError{
				Err: err,
				StatusCode: http.StatusUnauthorized,
			}
		}
		cardStripeId = card.StripeID
	} else {
		cardStripeId = data["stripe_card_id"].(string)
	}

	cusStripeId := data["stripe_cus_id"].(string)

	busStripeId := data["stripe_bus_id"].(string)


	stripeSubId, err := my_stripe.ResumeSubscription(
		cusStripeId,
		busStripeId,
		*plan.StripePriceID,
		cardStripeId,
		sub.Expires.Time,
	)


	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	fmt.Println(*stripeSubId)

	err = s.ResumeSubscription(subId, cardId, *stripeSubId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	return nil
}


func CancelSubscription(
	sqlDB *sql.DB,
	cusId int,
	subId int,
) (*time.Time, *models.RequestError) {
	s := db.SubscriptionDB{DB: sqlDB}

	// 1. check if cus owns sub
	fmt.Println(subId)
	fmt.Println("1. checking if cus owns sub")
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

	fmt.Println("1. deleting sub from stripe")

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
	expires := GetNextBillingDate(recurring, lastInvoice.Created)

	// // 3. update sql
	err = s.CancelSubscription(subId, expires)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	fmt.Println("Made it here:", expires)

	return &expires, nil
}

func ChangeSubDefaultCard(
	sqlDB *sql.DB,
	cusId int,
	subId int,
	cardId int,
) ( *models.RequestError) {
	s := db.SubscriptionDB{DB: sqlDB}
	c := db.CustomerDB{DB: sqlDB}

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
