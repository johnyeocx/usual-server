package business

import (
	"database/sql"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
)

func createSubProduct (
	sqlDB *sql.DB,
	businessId int,
	category *models.ProductCategory,
	product *models.Product, 
	subPlan *models.SubscriptionPlan,
) (*int, *models.SubscriptionProduct, error) {

	
	db := db.BusinessDB{DB: sqlDB}

	// 1. Insert new category id
	var newCatId *int;
	var catId = category.CategoryID
	
	if (category.CategoryID == nil) {
		id, err := db.InsertProductCategory(businessId, category.Title)
		if err != nil {
			return nil, nil, err
		}


		newCatId = id;
		catId = id;
	}

	stripeProductId, stripePriceId, err := my_stripe.CreateNewSubProduct(product.Name, *subPlan)
	if err != nil {
		return nil, nil, err
	}

	
	insertedProduct, err := db.InsertProduct(businessId, catId, product, *stripeProductId)
	if err != nil {
		return nil, nil, err
	}


	var insertedPlan *models.SubscriptionPlan
	if (subPlan.UsageUnlimited) {
		insertedPlan, err = db.InsertSubscriptionPlanUnlimitedUsage(insertedProduct.ProductID, subPlan, *stripePriceId)
	} else {
		insertedPlan, err = db.InsertSubscriptionPlanFiniteUsage(insertedProduct.ProductID, subPlan, *stripePriceId)
	}

	if err != nil {
		return nil, nil, err
	}
	

	return newCatId, &models.SubscriptionProduct{
		Product: *insertedProduct,
		SubPlan: *insertedPlan,
	}, nil
}

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

