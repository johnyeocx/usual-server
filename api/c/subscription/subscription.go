package subscription

import (
	"database/sql"
	"time"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
)


func CreateSubscription(
	sqlDB *sql.DB, 
	customerId int,
	productIds []int, 
) (error) {

	// Get list of products + subplans
	b := db.BusinessDB{DB: sqlDB}
	c := db.CustomerDB{DB: sqlDB}

	subProducts, err := b.GetSubProductsFromIds(productIds)
	if err != nil {
		return err
	}

	// Get stripe customer id
	stripeCustomerId, err := c.GetCustomerStripeId(customerId)
	if err != nil {
		return err
	}

	subscriptions := []models.Subscription{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// 3. INSERT INTO STRIPE
	for i := 0; i < len(*subProducts); i++ {
		stripeBusinessId, err := b.GetBusinessStripeID((*subProducts)[i].Product.BusinessID)
		if err != nil {
			return err
		}	

		stripeSubId, err := my_stripe.CreateSubscription(*stripeCustomerId, *stripeBusinessId, (*subProducts)[i])
		if err != nil {
			return err
		}
		subscription := models.Subscription{
			StripeSubID: *stripeSubId,
			CustomerID: customerId,
			PlanID: (*subProducts)[i].SubPlan.PlanID,
			StartDate: today,
		}
		subscriptions = append(subscriptions, subscription)
	}
	
	// 4. INSERT INTO DB
	s := db.SubscriptionDB{DB: sqlDB}
	err = s.InsertSubscriptions(&subscriptions)
	return err
}