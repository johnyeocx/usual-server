package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/johnyeocx/usual/server/db/models"
)

func (s *BusinessDB) GetBusinessSubProducts(
	businessId int,
) (*[]models.SubscriptionProduct, error) {

	selectStatement := `SELECT 
	product.product_id, business_id, name, description, category_id, stripe_product_id,
	plan_id, currency, recurring_interval, recurring_interval_count, unit_amount,

	usage_unlimited, usage_interval, usage_interval_count, usage_amount

	from product JOIN subscription_plan on product.product_id = subscription_plan.product_id
	WHERE business_id=$1 ORDER BY product.category_id ASC`

	// usage_amount
	rows, err := s.DB.Query(selectStatement, businessId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var subProducts []models.SubscriptionProduct
	for rows.Next() {
		var product models.Product
		var subPlan models.SubscriptionPlan

		var usageDurationInterval models.JsonNullString
		var usageDurationIntervalCount models.JsonNullInt16
		var usageAmount models.JsonNullInt16

        if err := rows.Scan(
			&product.ProductID,
			&product.BusinessID,
			&product.Name,
			&product.Description,
			&product.CategoryID,
			&product.StripeProductID,

			&subPlan.PlanID,
			&subPlan.Currency,
			&subPlan.RecurringDuration.Interval,
			&subPlan.RecurringDuration.IntervalCount,
			&subPlan.UnitAmount,
			&subPlan.UsageUnlimited,
			&usageDurationInterval,
			&usageDurationIntervalCount,
			&usageAmount,
		); err != nil {
            return &subProducts, err
        }
		
		if (!subPlan.UsageUnlimited) {
			subPlan.UsageDuration = &models.TimeFrame{
				Interval: usageDurationInterval,
				IntervalCount: usageDurationIntervalCount,
			}
			subPlan.UsageAmount = &usageAmount
		}
		
        subProducts = append(subProducts, models.SubscriptionProduct{
			Product: product,
			SubPlan: subPlan,
		})
    }



	return &subProducts, nil
}

func (s *BusinessDB) GetBusinessProductCategories(
	businessId int,
) (*[]models.ProductCategory, error){
	selectStatement := `SELECT 
	category_id, business_id, title 
	from product_category WHERE business_id=$1`

	rows, err := s.DB.Query(selectStatement, businessId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var categories []models.ProductCategory
	for rows.Next() {
		var productCat models.ProductCategory
        if err := rows.Scan(
			&productCat.CategoryID,
			&productCat.BusinessID,
			&productCat.Title,
		); err != nil {
            return &categories, err
        }
		
        categories = append(categories, productCat)
    }

	return &categories, nil
}

func (s *BusinessDB) GetSubProductsFromIds(
	productIds []int,
)(*[]models.SubscriptionProduct, error) {
	selectStatement := `SELECT 
	product.product_id, business_id, name, description, category_id, stripe_product_id,
	plan_id, currency, recurring_interval, recurring_interval_count, unit_amount, stripe_price_id

	from product JOIN subscription_plan on product.product_id = subscription_plan.product_id
	WHERE product.product_id=$1 ORDER BY business_id ASC`

	subProducts := []models.SubscriptionProduct{}
	// usage_amount
	for _, productId := range productIds{

		var product models.Product
		var subPlan models.SubscriptionPlan
		err := s.DB.QueryRow(selectStatement, productId).Scan(
			&product.ProductID,
			&product.BusinessID,
			&product.Name,
			&product.Description,
			&product.CategoryID,
			&product.StripeProductID,

			&subPlan.PlanID,
			&subPlan.Currency,
			&subPlan.RecurringDuration.Interval,
			&subPlan.RecurringDuration.IntervalCount,
			&subPlan.UnitAmount,
			&subPlan.StripePriceID,
		)

		if err != nil {
			continue
		}

		subProducts = append(subProducts, models.SubscriptionProduct{
			Product: product,
			SubPlan: subPlan,
		})
	
	}

	return &subProducts, nil
}

func (s *BusinessDB) GetSubProductSubscriptions(
	productId int,
) () {

}

func (s *BusinessDB) InsertProductCategory(
	businessId int, 
	title string,
) (*int, error) {
	var categoryId int;
	err := s.DB.QueryRow(`INSERT into 
		product_category (business_id, title) 
		VALUES ($1, $2) RETURNING category_id`, 
		businessId, 
		title,
	).Scan(&categoryId)

	if err != nil {
		return nil, err
	}
	return &categoryId, nil
}

func (s *BusinessDB) InsertProduct(
	businessId int, 
	categoryId *int, 
	product *models.Product,
	stripeProductId string,
) (*models.Product, error) {
	
	var insertedProduct models.Product;

	err := s.DB.QueryRow(`INSERT into 
		product (business_id, name, description, category_id, stripe_product_id) 
		VALUES ($1, $2, $3, $4, $5) RETURNING product_id, business_id, name, description, category_id, stripe_product_id`, 
		businessId, 
		product.Name,
		product.Description,
		categoryId,
		stripeProductId,
	).Scan(
		&insertedProduct.ProductID,
		&product.BusinessID,
		&insertedProduct.Name,
		&insertedProduct.Description,
		&insertedProduct.CategoryID,
		&insertedProduct.StripeProductID,
	)

	if err != nil {
		return nil, err
	}
	return &insertedProduct, nil
}


func (s *BusinessDB) InsertSubscriptionPlanUnlimitedUsage(
	productId int, 
	subscription *models.SubscriptionPlan,
	stripePriceId string,
) (*models.SubscriptionPlan, error) {
	
	var plan models.SubscriptionPlan;

	err := s.DB.QueryRow(`INSERT into 
		subscription_plan (product_id, currency,
			recurring_interval, recurring_interval_count, 
			unit_amount, usage_unlimited, stripe_price_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING 
		plan_id, product_id, currency, recurring_interval, recurring_interval_count,
		unit_amount, usage_unlimited, stripe_price_id
		`, 
		
		productId, subscription.Currency,
		subscription.RecurringDuration.Interval, subscription.RecurringDuration.IntervalCount,
		subscription.UnitAmount, subscription.UsageUnlimited, stripePriceId,
	).Scan(
		&plan.PlanID,
		&plan.ProductID,
		&plan.Currency,
		&plan.RecurringDuration.Interval,
		&plan.RecurringDuration.IntervalCount,
		&plan.UnitAmount,
		&plan.UsageUnlimited,
		&plan.StripePriceID,
	)

	if err != nil {
		log.Print(err)
		return nil, err
	}
	
	return &plan, nil
}


func (s *BusinessDB) InsertSubscriptionPlanFiniteUsage(
	productId int, 
	subscription *models.SubscriptionPlan,
	stripePriceId string,
) (*models.SubscriptionPlan, error) {

	var plan models.SubscriptionPlan;

	var insertedUsageDuration models.TimeFrame
	var usageAmount models.JsonNullInt16;

	err := s.DB.QueryRow(`INSERT into 
		subscription_plan (product_id, currency,
			recurring_interval, recurring_interval_count, 
			unit_amount, usage_unlimited,
			usage_interval, usage_interval_count, 
			usage_amount, stripe_price_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING 
		plan_id, product_id, currency, recurring_interval, recurring_interval_count,
		unit_amount, usage_unlimited, usage_interval, usage_interval_count, usage_amount, stripe_price_id
		`, 
		
		productId, subscription.Currency,
		subscription.RecurringDuration.Interval, subscription.RecurringDuration.IntervalCount,
		subscription.UnitAmount, subscription.UsageUnlimited,
		subscription.UsageDuration.Interval, subscription.UsageDuration.IntervalCount,
		*subscription.UsageAmount, stripePriceId,

	).Scan(
		&plan.PlanID,
		&plan.ProductID,
		&plan.Currency,
		&plan.RecurringDuration.Interval,
		&plan.RecurringDuration.IntervalCount,
		&plan.UnitAmount,
		&plan.UsageUnlimited,
		&insertedUsageDuration.Interval,
		&insertedUsageDuration.IntervalCount,
		&usageAmount,
		&plan.StripePriceID,
	)

	if err != nil {
		log.Print(err)
		return nil, err
	}

	plan.UsageDuration = &insertedUsageDuration
	plan.UsageAmount = &usageAmount
	
	return &plan, nil
}


// PATCH, EDITS
func (s *BusinessDB) BusinessOwnsProductAndPlan(
	businessId int,
	productId int,
	planId int,
) (error) {

	// CHECK THAT PRODUCT & PLAN BELONGS TO BUSINESS ID
	var businessMatch bool
	err := s.DB.QueryRow(`
		SELECT business_id=$1 
		from product JOIN subscription_plan ON product.product_id=subscription_plan.plan_id
		WHERE product.product_id=$2 AND subscription_plan.plan_id=$3`, 
		businessId, productId, planId,
	).Scan(&businessMatch)

	return err
}

func (s *BusinessDB) BusinessOwnsProduct(
	businessId int,
	productId int,
) (error) {

	// CHECK THAT PRODUCT & PLAN BELONGS TO BUSINESS ID
	var businessMatch bool
	err := s.DB.QueryRow(`
		SELECT business_id=$1 
		from product WHERE product_id=$2`, 
		businessId, productId,
	).Scan(&businessMatch)

	return err
}


func (s *BusinessDB) GetStripePriceId(
	planId int,
) (*string, error) {

	// CHECK THAT PRODUCT BELONGS TO BUSINESS ID
	var stripePriceId string
	err := s.DB.QueryRow(`SELECT stripe_price_id from subscription_plan WHERE plan_id=$1`, 
		planId,
	).Scan(&stripePriceId)
	
	return &stripePriceId, err
}

func (s *BusinessDB) GetStripeProductId(
	productId int,
) (*string, error) {

	// CHECK THAT PRODUCT BELONGS TO BUSINESS ID
	var stripeProductId string
	err := s.DB.QueryRow(`SELECT stripe_product_id from product WHERE product_id=$1`, 
		productId,
	).Scan(&stripeProductId)
	
	return &stripeProductId, err
}

func (s *BusinessDB) SetProductName(
	businessId int,
	productId int,
	name string,
) (error) {
	_, err := s.DB.Exec(`UPDATE product SET name=$1 WHERE product_id=$2 AND business_id=$3`, 
		name, productId, businessId,
	)

	return err
}

func (s *BusinessDB) SetProductDescription(
	businessId int,
	productId int,
	description string,
) (error) {
	_, err := s.DB.Exec(`UPDATE product SET description=$1 WHERE product_id=$2 AND business_id=$3`, 
		description, productId, businessId,
	)

	return err
}

func (s *BusinessDB) SetProductCategory(
	businessId int,
	productId int,
	categoryId *int,
	title string,
) (*int, error) {
	
	// CREATE NEW CATEGORY IF ID IS NULL
	var finalCatId int;
	if (categoryId == nil) {
		err := s.DB.QueryRow(`INSERT into product_category (business_id, title) VALUES ($1, $2) RETURNING category_id`, 
			businessId, title,
		).Scan(&finalCatId)

		if err != nil {
			return nil, err
		}
	} else {
		finalCatId = *categoryId
	}

	
	
	// UPDATE PRODUCT CATEGORY
	_, err := s.DB.Exec(`UPDATE product SET category_id=$1 WHERE product_id=$2 AND business_id=$3`, 
		finalCatId, productId, businessId,
	)
	
	if err != nil {
		return nil, err
	}

	if categoryId == nil {

		return &finalCatId, nil
	}
	return nil, nil
}

func (s *BusinessDB) SetSubProductPricing(
	businessId int,
	productId int,
	planId int,
	recurringDuration models.TimeFrame,
	unitAmount int,
	stripePriceId string,
) (error) {
	
	// CHECK THAT PRODUCT BELONGS TO BUSINESS ID
	var businessMatch bool
	err := s.DB.QueryRow(`SELECT business_id=$1 from product WHERE product_id=$2`, 
		businessId, productId,
	).Scan(&businessMatch)

	if err != nil {
		return err
	}

	if !businessMatch {
		return errors.New("product id does not belong to business id")
	}
	fmt.Println("Stripe price id:", stripePriceId)

	_, err = s.DB.Exec(`UPDATE subscription_plan SET 
		recurring_interval=$1, recurring_interval_count=$2, unit_amount=$3, stripe_price_id=$4
		WHERE product_id=$5 AND plan_id=$6`, 
	 	recurringDuration.Interval.String, 
		recurringDuration.IntervalCount.Int16, 
		unitAmount, 
		stripePriceId,
		productId, 
		planId, 
	)

	return err
}

func (s *BusinessDB) SetSubProductUsage(
	businessId int,
	productId int,
	planId int,
	usageUnlimited bool, 
	usageDuration *models.TimeFrame,
	usageAmount *int,
) (error) {
	
	// CHECK THAT PRODUCT BELONGS TO BUSINESS ID
	var businessMatch bool
	err := s.DB.QueryRow(`SELECT business_id=$1 from product WHERE product_id=$2`, 
		businessId, productId,
	).Scan(&businessMatch)

	if err != nil {
		return err
	}

	if !businessMatch {
		return errors.New("product id does not belong to business id")
	}
	

	if (usageUnlimited) {
		nullInt := sql.NullInt16{Valid: false}
		nullString := sql.NullString{Valid: false}
		_, err = s.DB.Exec(`UPDATE subscription_plan SET 
			usage_unlimited=$1, usage_interval=$2, usage_interval_count=$3, usage_amount=$4
			WHERE product_id=$5 AND plan_id=$6`, 
			usageUnlimited, nullString, nullInt, nullInt, productId, planId,
		)

		if err != nil {
			return err
		}
	} else {
		_, err = s.DB.Exec(`UPDATE subscription_plan SET 
			usage_unlimited=$1, usage_interval=$2, usage_interval_count=$3, usage_amount=$4
			WHERE product_id=$5 AND plan_id=$6`, 
			usageUnlimited, 
			usageDuration.Interval.String, usageDuration.IntervalCount.Int16, usageAmount, 
			productId, planId,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

