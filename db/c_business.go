package db

import (
	"database/sql"

	"github.com/johnyeocx/usual/server/db/models"
)

type BusinessDB struct {
	DB *sql.DB
}

func (b *BusinessDB) GetBusinessWithTopSubbedProduct(
) ([]models.ExploreResult, error){
	query := `
	WITH ranked_table AS (

		SELECT b.business_id, sp.plan_id, COUNT(s.plan_id) as sub_count, ROW_NUMBER() OVER 
		(PARTITION BY b.business_id ORDER BY COUNT(s.plan_id) DESC) as rank
		FROM business as b 
		JOIN product as p ON b.business_id=p.business_id
		JOIN subscription_plan as sp on sp.product_id=p.product_id
		LEFT JOIN subscription as s ON sp.plan_id = s.plan_id
		GROUP BY b.business_id, sp.plan_id
		ORDER BY sub_count DESC, sp.plan_id ASC
	)
	
	SELECT 
	b.business_id, b.name, b.business_category, b.description, b.business_url,
	p.product_id, p.name, p.description,
	sp.plan_id, recurring_interval, recurring_interval_count, unit_amount, sub_count,
	pc.title
	FROM ranked_table as r 
	JOIN business as b ON r.business_id=b.business_id
	JOIN subscription_plan as sp ON r.plan_id=sp.plan_id
	JOIN product as p on sp.product_id=p.product_id
	JOIN product_category as pc on p.category_id=pc.category_id
	WHERE rank = 1
	`

	rows, err := b.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := []models.ExploreResult{}

	for rows.Next() {
		var product models.Product
		var business models.Business
		var plan models.SubscriptionPlan

        if err := rows.Scan(
			&business.ID,
			&business.Name,
			&business.BusinessCategory,
			&business.Description,
			&business.BusinessUrl,
			&product.ProductID,
			&product.Name,
			&product.Description,
			&plan.PlanID,
			&plan.RecurringDuration.Interval,
			&plan.RecurringDuration.IntervalCount,
			&plan.UnitAmount,
			&product.SubCount,
			&product.CatTitle,
		); err != nil {
			continue
        }

		results = append(results, models.ExploreResult{
			Business: business,
			SubProduct: models.SubscriptionProduct{
				Product: product,
				SubPlan: plan,
			},
		})
    }

	return results, nil
}

func (b *BusinessDB) SearchSubProducts(q string) ([]models.ExploreResult, error){
		search := "%" + q + "%"
		query := `
		SELECT 
		b.business_id, b.name, b.business_category, b.description, b.business_url,
		p.product_id, p.name, p.description, pc.title,
		sp.plan_id, recurring_interval, recurring_interval_count, unit_amount
				
		FROM product as p 
		JOIN subscription_plan as sp on sp.product_id=p.product_id 
		JOIN product_category as pc on p.category_id=pc.category_id
		JOIN business as b on b.business_id=p.business_id 
		
		WHERE
		LOWER(pc.title) LIKE $1 OR
		LOWER(p.name) LIKE $1 OR
		LOWER(p.description) LIKE $1
		
		GROUP BY p.product_id, b.business_id, pc.category_id, sp.plan_id
		
		order by
		case
			WHEN LOWER(p."name") LIKE $1 then 0
			WHEN LOWER(pc.title) LIKE $1 then 1
			WHEN LOWER(p.description) LIKE $1 then 2
			else 3
		end asc;
		`
	
		rows, err := b.DB.Query(query, search)
		if err != nil {
			return nil, err
		}
	
		defer rows.Close()
	
		results := []models.ExploreResult{}
	
		for rows.Next() {
			var product models.Product
			var business models.Business
			var plan models.SubscriptionPlan
	
			if err := rows.Scan(
				&business.ID,
				&business.Name,
				&business.BusinessCategory,
				&business.Description,
				&business.BusinessUrl,
				&product.ProductID,
				&product.Name,
				&product.Description,
				&product.CatTitle,
				&plan.PlanID,
				&plan.RecurringDuration.Interval,
				&plan.RecurringDuration.IntervalCount,
				&plan.UnitAmount,
			); err != nil {
				continue
			}
	
			results = append(results, models.ExploreResult{
				Business: business,
				SubProduct: models.SubscriptionProduct{
					Product: product,
					SubPlan: plan,
				},
			})
		}
	
		return results, nil
	}

func (b *BusinessDB) SearchAccounts(
	q string,	
) ([]models.Business, error) {
	search := "%" + q + "%"
	query := `
	SELECT b.business_id, b.name, b.business_category, b.description, b.business_url from business as b 
	JOIN product_category as pc on b.business_id=pc.business_id
	JOIN product as p on p.business_id=b.business_id WHERE
	LOWER(b.name) LIKE $1 OR
	LOWER(b.business_category) LIKE $1 OR
	LOWER(pc.title) LIKE $1 OR
	LOWER(p.name) LIKE $1 OR
	LOWER(p.description) LIKE $1
	
	GROUP BY b.business_id
	
	order by
		case
			WHEN LOWER(b.name) LIKE $1 then 0
			WHEN LOWER(b.business_category) LIKE $1 then 1
			else 2
		end asc;
	`

	rows, err := b.DB.Query(query, search)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	accounts := []models.Business{}

	for rows.Next() {
		var business models.Business

        if err := rows.Scan(
			&business.ID,
			&business.Name,
			&business.BusinessCategory,
			&business.Description,
			&business.BusinessUrl,
		); err != nil {
			continue
        }

		accounts = append(accounts, business)
    }

	return accounts, nil
}

func (s *BusinessDB) GetCBusinessSubProducts(
	businessId int,
) (*[]models.SubscriptionProduct, error) {

	selectStatement := `SELECT 
	p.product_id, p.name, p.description, p.category_id, sp.plan_id, sp.currency, 
	recurring_interval, recurring_interval_count, unit_amount, COUNT(DISTINCT c.customer_id) as sub_count

	from product as p
	JOIN subscription_plan as sp on p.product_id = sp.product_id
	LEFT JOIN subscription as s on s.plan_id=sp.plan_id
	LEFT JOIN customer as c on c.customer_id=s.customer_id
	WHERE business_id=$1 

	GROUP BY p.product_id, sp.plan_id
	ORDER BY p.product_id ASC`

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

        if err := rows.Scan(
			&product.ProductID,
			&product.Name,
			&product.Description,
			&product.CategoryID,
			&subPlan.PlanID,
			&subPlan.Currency,
			&subPlan.RecurringDuration.Interval,
			&subPlan.RecurringDuration.IntervalCount,
			&subPlan.UnitAmount,
			&product.SubCount,
		); err != nil {
            return &subProducts, err
        }
	
		
        subProducts = append(subProducts, models.SubscriptionProduct{
			Product: product,
			SubPlan: subPlan,
		})
    }

	return &subProducts, nil
}

func (s *BusinessDB) GetCSubProduct(
	productId int,
) (*models.SubscriptionProduct, error) {

	stmt := `SELECT 
	p.product_id, p.name, p.description, p.category_id, p.business_id, sp.plan_id, sp.currency, 
	recurring_interval, recurring_interval_count, unit_amount, pc.title
	
	from product as p
	JOIN subscription_plan as sp on p.product_id = sp.product_id
	JOIN product_category as pc on pc.category_id=p.category_id
	WHERE p.product_id=$1
	`

	product := models.Product{}
	plan := models.SubscriptionPlan{}
	plan.RecurringDuration = models.TimeFrame{}

	err := s.DB.QueryRow(stmt, productId).Scan(
		&product.ProductID,
		&product.Name,
		&product.Description,
		&product.CategoryID,
		&product.BusinessID,
		&plan.PlanID,
		&plan.Currency,
		&plan.RecurringDuration.Interval,
		&plan.RecurringDuration.IntervalCount,
		&plan.UnitAmount,
		&product.CatTitle,
	)
	if err != nil {
		return nil, err
	}

	subProduct := models.SubscriptionProduct{
		Product: product,
		SubPlan: plan,
	}
	return &subProduct , nil
}

func (businessDB *BusinessDB) GetBusinessByIDWithSubCount(businessId int) (*models.Business, error) {

	selectStatement := `
	SELECT COUNT (DISTINCT c.customer_id) as sub_count, b.business_id, b.name, b.email, b.country, b.business_category, b.business_url, b.description
	FROM business as b 
	JOIN product as p on b.business_id=p.business_id
	JOIN subscription_plan as sp on sp.product_id=p.product_id
	LEFT JOIN subscription as s on sp.plan_id=s.plan_id
	LEFT JOIN customer as c on c.customer_id=s.customer_id
	WHERE b.business_id=$1
	
	GROUP BY b.business_id`

	var business models.Business
	if err := businessDB.DB.QueryRow(selectStatement, businessId).Scan(
		&business.SubCount,
		&business.ID,
		&business.Name,
		&business.Email,
		&business.Country,
		&business.BusinessCategory,
		&business.BusinessUrl,
		&business.Description,
	); err != nil {
		return nil, err
	}

	return &business, nil
}