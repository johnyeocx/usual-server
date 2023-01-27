package business

import (
	"database/sql"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
)


func GetBusinessProducts(
	sqlDB *sql.DB,
	businessId int,
) (*[]models.ProductCategory, *[]models.SubscriptionProduct, error) {

	b := db.BusinessDB{DB: sqlDB}
	productCategories, err := b.GetBusinessProductCategories(businessId)
	if err != nil {
		return nil, nil,err
	}

	subProducts, err := b.GetBusinessSubProducts(businessId)
	if err != nil {
		return nil,nil, err
	}

	return productCategories, subProducts, nil
}

func UpdateSubProductPricing(
	sqlDB *sql.DB,
	businessId int,
	productId int,
	planId int,
	recurringDuration models.TimeFrame,
	unitAmount int,
) (error) {
	b := db.BusinessDB{DB: sqlDB}

	// 1. check business owns product
	err := b.BusinessOwnsProductAndPlan(businessId, productId, planId)
	if err != nil {
		return err
	}

	// 2. Get stripe price id
	stripePriceId, err := b.GetStripePriceId(planId)
	if err != nil {
		return err
	}
	stripeProductId, err := b.GetStripeProductId(productId)
	if err != nil {
		return err
	}
	
	
	newPriceId , err:= my_stripe.UpdateSubProductPrice(*stripePriceId, *stripeProductId, recurringDuration, unitAmount)
	if err != nil {
		return err
	}

	err = b.SetSubProductPricing(businessId, productId, planId, recurringDuration, unitAmount, *newPriceId)
	if err != nil {
		return err
	}

	return nil
}

