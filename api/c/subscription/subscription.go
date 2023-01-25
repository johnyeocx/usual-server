package subscription

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
)


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

func CancelSubscription(
	sqlDB *sql.DB,
	cusId int,
	subId int,
) (*models.RequestError) {
	s := db.SubscriptionDB{DB: sqlDB}

	// 1. check if cus owns sub
	sub, err := s.CusOwnsSub(cusId, subId)
	if err != nil && err != sql.ErrNoRows {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	if err == sql.ErrNoRows {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadRequest,
		}
	}

	// 2. delete from stripe
	err = my_stripe.CancelSubscription(sub.StripeSubID)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 3. delete from sql
	err = s.DeleteSubscription(subId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	fmt.Println("Completed")

	return nil
}