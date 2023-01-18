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

	// if err == sql.ErrNoRows {
	// 	empty := []models.SubscriptionProduct{}
	// 	return &empty, nil
	// }
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

func (s *BusinessDB) GetSubProductDeleteData(
	productId int,
) (*map[string]interface{}, error) {
	stmt1 := `SELECT p.stripe_product_id, p.category_id, sp.plan_id, sp.stripe_price_id FROM 
	product as p JOIN subscription_plan as sp ON p.product_id=sp.product_id WHERE p.product_id=$1;`

	
	var stripeProductId string;
	var catId int;
	var stripePriceId string;
	var planId string;
	if err := s.DB.QueryRow(stmt1, productId).Scan(
		&stripeProductId,
		&catId,
		&planId,
		&stripePriceId,
	); err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"stripe_product_id": stripeProductId,
		"category_id": catId,
		"stripe_price_id": stripePriceId,
		"plan_id": planId,
	}
	
	stmt2 := `SELECT stripe_sub_id FROM subscription WHERE plan_id=$1`
	rows, err := s.DB.Query(stmt2, planId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	subIds := []string{}
	for rows.Next() {
		var subId string
        if err := rows.Scan(
			&subId,
		); err != nil {
			continue
        }
		subIds = append(subIds, subId)
    }

	data["stripe_sub_ids"] = subIds

	return &data, nil
}


// Sub Product Stats

func (s *BusinessDB) GetSubProductInvoices(
	productId int,
) (*[]models.InvoiceData, error) {
	// stmt := `SELECT * from product p JOIN subscription_plan sp JOIN subscription s
	// 	ON p.product_id=sp.product_id AND sp.plan_id=s.plan_id`
	
	stmt := `SELECT 
	c.customer_id, c.stripe_id, c.name, i.invoice_id, i.total, i.invoice_url, i.status, i.attempted, i.app_fee_amt from 
	product as p JOIN invoice as i ON p.stripe_product_id=i.stripe_prod_id
	JOIN customer as c on i.stripe_cus_id=c.stripe_id
	WHERE p.product_id=$1;`

	rows, err := s.DB.Query(stmt, productId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	invoices := []models.InvoiceData{}
	for rows.Next() {
		var invoice models.InvoiceData
		
        if err := rows.Scan(
			&invoice.CustomerID,
			&invoice.CustomerStripeID,
			&invoice.CustomerName,
			&invoice.InvoiceID,
			&invoice.Total,
			&invoice.InvoiceURL,
			&invoice.Status,
			&invoice.Attempted,
			&invoice.ApplicationFeeAmt,
		); err != nil {
			fmt.Println(err)
			continue
        }
		
		invoices = append(invoices, invoice)
    }

	return &invoices, nil
}

func (s *BusinessDB) GetSubProductSubscribers(
	productId int,
) (*[]models.Customer, error) {
	stmt := `SELECT c.customer_id, c.name  from product as 
	p JOIN subscription_plan as sp ON p.product_id=sp.product_id
	JOIN subscription as s ON s.plan_id=sp.plan_id 
	JOIN customer as c ON s.customer_id=c.customer_id
	WHERE p.product_id=$1;`

	
	rows, err := s.DB.Query(stmt, productId)
	if err != nil {
		return nil, err
	}
	

	defer rows.Close()

	subscribers := []models.Customer{}
	for rows.Next() {
		var subscriber models.Customer
        if err := rows.Scan(
			&subscriber.ID,
			&subscriber.Name,
		); err != nil {
			continue
        }
		
		subscribers = append(subscribers, subscriber)
    }

	
	return &subscribers, nil
}



// INSERTS
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

func (s *BusinessDB) DeleteSubscription(
	stripeSubId string,
) (error) {

	stmt1 := `DELETE from subscription WHERE stripe_sub_id=$1`
	_, err := s.DB.Exec(stmt1, stripeSubId)
	return err
}

func (s *BusinessDB) DeleteProductAndPlan(
	productId int,
) (error) {

	_, err := s.DB.Exec(`DELETE from subscription_plan WHERE product_id=$1`, productId)
	if err != nil {
		return err
	}
	
	_, err = s.DB.Exec(`DELETE from product WHERE product_id=$1`, productId)
	return err
}

func (s *BusinessDB) DeleteCategoryIfEmpty(
	categoryId int,
)(error) {

	var pId int
	err := s.DB.QueryRow(`
		SELECT product_id from product WHERE category_id=$1`, 
		categoryId).Scan(&pId)

	if err == sql.ErrNoRows {
		_, err = s.DB.Exec(`DELETE from product_category WHERE category_id=$1`, categoryId)
		return err
	}

	return nil
}