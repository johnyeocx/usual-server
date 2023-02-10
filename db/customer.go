package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/secure"
)

type CustomerDB struct {
	DB	*sql.DB
}

func (c *CustomerDB) CreateCustomer (
	name string,
	email string,
	password string,
	uuid string,
) (*int, error) {
	hashedPassword, err := secure.GenerateHashFromStr(password)
	if err != nil {
		return nil, err
	}

	var cusId int
	err = c.DB.QueryRow(`
		INSERT into customer (name, email, password, uuid) VALUES ($1, $2, $3, $4) RETURNING customer_id`,
		name, email, hashedPassword, uuid,
	).Scan(&cusId)

	if err != nil {
		return nil, err
	}

	return &cusId, nil
}

func (c *CustomerDB) CreateCFromSubscribe (
	name string,
	email string,
	stripeId string,
) (*int, error) {
	var cId int
	err := c.DB.QueryRow(`INSERT into customer (name, email, stripe_id)
		VALUES ($1, $2, $3) RETURNING customer_id`,	
		name, email, stripeId,
	).Scan(&cId)

	return &cId, err
}

func (c *CustomerDB) InsertCustomerStripeID (
	cusId int,
	stripeId string,
) (error) {
	_, err := c.DB.Exec(`UPDATE customer SET stripe_id=$1 WHERE customer_id=$2`,
		stripeId, cusId,
	)

	return err
}

func (c *CustomerDB) GetCustomerEmailVerified (
	email string,
) (bool, error) {
	
	var verified bool
	err := c.DB.QueryRow("SELECT email_verified FROM customer WHERE email=$1", 
	email).Scan(&verified)

	return verified, err
}

func (c *CustomerDB) GetCustomerByID (
	cusId int,
) (*models.Customer, error) {
	var cus models.Customer
	cus.Address = &models.CusAddress{} 


	err := c.DB.QueryRow(`SELECT 
	c.customer_id, c.name, c.email, c.stripe_id, c.default_card_id, 
	c.address_line1, c.address_line2, c.postal_code, c.city, c.country
	FROM customer as c 
	WHERE customer_id=$1
	GROUP BY c.customer_id`, 
	cusId).Scan(
		&cus.ID,
		&cus.Name,
		&cus.Email,
		&cus.StripeID,
		&cus.DefaultCardID,
		&cus.Address.Line1,
		&cus.Address.Line2,
		&cus.Address.PostalCode,
		&cus.Address.City,
		&cus.Address.Country,
	)
	if err != nil {
		return nil, err
	}

	return &cus, nil
}

func (c *CustomerDB) GetCustomerWithTotalByID (
	cusId int,
) (*models.Customer, *int, error) {
	var cus models.Customer
	cus.Address = &models.CusAddress{} 

	var total sql.NullInt64

	now := time.Now()
	monthAgo := time.Date(now.Year(), now.Month() - 1, now.Day(), 0, 0, 0, 0, time.UTC)

	err := c.DB.QueryRow(`SELECT 
	c.customer_id, c.name, c.email, c.stripe_id, c.default_card_id, c.address_line1, c.address_line2, c.postal_code, c.city, c.country,
	SUM(i.total) as total
	FROM customer as c 
	LEFT JOIN invoice as i on i.stripe_cus_id=c.stripe_id AND i.created > $1
	WHERE customer_id=$2
	GROUP BY c.customer_id, i.invoice_id`, 
	monthAgo, cusId).Scan(
		&cus.ID,
		&cus.Name,
		&cus.Email,
		&cus.StripeID,
		&cus.DefaultCardID,
		&cus.Address.Line1,
		&cus.Address.Line2,
		&cus.Address.PostalCode,
		&cus.Address.City,
		&cus.Address.Country,
		&total,
	)

	if err != nil {
		return nil, nil, err
	}

	totalInt := 0
	if total.Valid {
		totalInt = int(total.Int64)	
	} 

	return &cus, &totalInt, nil
}

func (c *CustomerDB) GetCustomerByEmail (
	email string,
) (*models.Customer, error) {
	var cus models.Customer 
	err := c.DB.QueryRow(`SELECT 
		customer_id, name, email, uuid
		FROM customer WHERE email=$1`, 
	email).Scan(
		&cus.ID,
		&cus.Name,
		&cus.Email,
		&cus.Uuid,
	)

	if err != nil {
		return nil, err
	}

	return &cus, nil
}

func (c *CustomerDB) GetCustomerStripeId (
	customerId int,
) (*string, error) {
	var stripeId string
	err := c.DB.QueryRow("SELECT stripe_id FROM customer WHERE customer_id=$1", 
		customerId,
	).Scan(&stripeId)

	if err != nil {
		return nil, err
	}
	return &stripeId, nil
}

func (c *CustomerDB) GetCustomerAndCardStripeId (
	cusId int,
	cardId int,
) (*string, *string, error) {

	var cusStripeId string
	var cardStripeId string
	err := c.DB.QueryRow(`SELECT c.stripe_id,cc.stripe_id FROM customer as c JOIN customer_card as cc ON c.customer_id=cc.customer_id
	WHERE c.customer_id=$1 AND cc.card_id=$2`, cusId, cardId).Scan(
		&cusStripeId,
		&cardStripeId,
	)

	if err != nil {
		return nil, nil, err
	}
	return &cusStripeId, &cardStripeId, nil
}

func (c *CustomerDB) GetCusPasswordFromEmail (
	email string,
) (*int, *string, error) {
	var cusId int
	var password string
	err := c.DB.QueryRow("SELECT customer_id, password FROM customer WHERE email=$1", 
		email,
	).Scan(&cusId, &password)

	if err != nil {
		return nil, nil, err
	}
	return &cusId, &password, nil
}

func (c *CustomerDB) GetCusPasswordFromID (
	cusId int,
) (*string, error) {

	var password string
	err := c.DB.QueryRow("SELECT password FROM customer WHERE customer_id=$1", 
		cusId,
	).Scan(&password)

	if err != nil {
		return nil, err
	}
	return &password, nil
}


func (c *CustomerDB) CheckCusSubscribed(cusId int, productIds []int) ( error) {
	
	valString := ""
	valueArgs := make([]interface{}, 0, len(productIds))
	valueArgs = append(valueArgs, cusId)

	for i, id := range productIds {
		if i == len(productIds) - 1 {
			valString += fmt.Sprintf("p.product_id=$%d)", i + 2)
		} else {
			valString += fmt.Sprintf("p.product_id=$%d OR", i + 2)
		}
		
		valueArgs = append(valueArgs, id)
	}

	query := fmt.Sprintf(`SELECT 
	p.product_id
	from customer as c
	JOIN subscription as s on c.customer_id=s.customer_id
	JOIN subscription_plan as sp on s.plan_id=sp.plan_id
	JOIN product as p on sp.product_id=p.product_id
	WHERE c.customer_id=$1 AND (%s`, valString)
	
	
	rows, err := c.DB.Query(query, valueArgs...)
	if err != nil {
		return err
	}

	if rows.Next() {
		return errors.New("customer is subscribed to one or more products in argument")
	}

	return nil
}

func (c *CustomerDB) GetCustomerSubscriptions(cusId int) ([]models.Subscription, error) {
	query := `
	WITH ranked_table AS (

		SELECT 
		c.customer_id,
		s.sub_id, s.start_date, s.cancelled, s.expires, s.cancelled_date, s.card_id,
		b.name, b.business_id,
		p.product_id, p.name, p.description, pc.title,
		sp.plan_id, sp.recurring_interval, sp.recurring_interval_count, sp.unit_amount, sp.currency,
		i.created, i.status, i.total,
		ROW_NUMBER() OVER 
		(PARTITION BY s.sub_id ORDER BY i.created DESC) as rank
		FROM customer as c 
		JOIN subscription as s ON c.customer_id=s.customer_id
		JOIN subscription_plan as sp ON sp.plan_id=s.plan_id
		JOIN product as p ON p.product_id=sp.product_id
		JOIN product_category as pc ON pc.category_id=p.category_id
		JOIN business as b ON b.business_id=p.business_id
		JOIN invoice as i ON i.stripe_prod_id=p.stripe_product_id
		GROUP BY sp.plan_id, p.product_id, s.sub_id, i.invoice_id, c.customer_id, pc.category_id, b.business_id
	)
	
	SELECT * FROM ranked_table as r
	WHERE r.customer_id=$1 AND rank=1
	`
	
	
	rows, err := c.DB.Query(query, cusId)
	if err != nil {
		return nil, err
	}

	subs := []models.Subscription{}

	for rows.Next() {
		var sub models.Subscription
		var product models.Product
		var plan models.SubscriptionPlan
		plan.RecurringDuration = models.TimeFrame{}
		var invoice models.Invoice
		
		var cusIdFiller int
		var rank int
		if err := rows.Scan(
			&cusIdFiller,
			&sub.ID, &sub.StartDate, &sub.Cancelled, &sub.Expires, &sub.CancelledDate, &sub.CardID,
			&sub.BusinessName, &sub.BusinessID,
			&product.ProductID, &product.Name, &product.Description, &product.CatTitle,
			&plan.PlanID, &plan.RecurringDuration.Interval, &plan.RecurringDuration.IntervalCount, &plan.UnitAmount, &plan.Currency,
			&invoice.Created, &invoice.Status, &invoice.Total, &rank,
		); err != nil {
			return nil, err
		}

		sub.SubProduct = &models.SubscriptionProduct{
			Product: product,
			SubPlan: plan,
		}
		sub.LastInvoice = &invoice
		subs = append(subs, sub)
	}

	return subs, nil
}

func (c *CustomerDB) GetCustomerCards(cusId int) ([]models.CardInfo, error) {
	query := `SELECT 
		cc.card_id, cc.last4, cc.brand FROM customer as c JOIN customer_card as cc on c.customer_id=cc.customer_id
		WHERE c.customer_id=$1
	`
	
	rows, err := c.DB.Query(query, cusId)
	if err != nil {
		return nil, err
	}

	cards := []models.CardInfo{}
	for rows.Next() {
		var card models.CardInfo
		if err := rows.Scan(
			&card.ID,
			&card.Last4,
			&card.Brand,
			
		); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (c *CustomerDB) GetCustomerInvoices(cusId int) ([]models.Invoice, error) {
	query := `
	SELECT 
	i.invoice_id, i.paid, i.attempted, i.status, i.total, i.created, i.invoice_url, 
	s.sub_id, s.plan_id, s.start_date, 
	p.product_id, p.name, b.business_id, b.name,
	cc.card_id, cc.brand, cc.last4
	from invoice as i
	JOIN customer as c ON i.stripe_cus_id=c.stripe_id
	JOIN subscription as s on i.sub_id=s.sub_id
	JOIN customer_card as cc on cc.card_id=s.card_id
	JOIN subscription_plan as sp on i.stripe_price_id=sp.stripe_price_id
	JOIN product as p on p.product_id=sp.product_id
	JOIN business as b on b.business_id=p.business_id
	

	WHERE c.customer_id=$1
	ORDER BY created DESC
	LIMIT 100
	`


	rows, err := c.DB.Query(query, cusId)
	if err != nil {
		return nil, err
	}

	invoices := []models.Invoice{}
	for rows.Next() {
		var in models.Invoice
		in.Subscription = &models.Subscription{}
		var product models.Product
		in.CardInfo = &models.CardInfo{}
		if err := rows.Scan(
			&in.ID, &in.Paid, &in.Attempted, &in.Status, &in.Total, &in.Created, &in.InvoiceURL,
			&in.Subscription.ID, &in.Subscription.PlanID, &in.Subscription.StartDate,
			&product.ProductID, &product.Name, &in.Subscription.BusinessID, &in.Subscription.BusinessName,
			&in.CardInfo.ID, &in.CardInfo.Brand, &in.CardInfo.Last4,
		); err != nil {
			return nil, err
		}
		in.Subscription.SubProduct = &models.SubscriptionProduct{
			Product: product,
		}
		invoices = append(invoices, in)
	}

	return invoices, nil
}

func (c *CustomerDB) GetTotalSpent(
	cusId int,
	productId int,
) (*int, error) {
	query := `
	SELECT 
	SUM(i.total)
	FROM customer as c 
	JOIN subscription as s on s.customer_id=c.customer_id
	JOIN subscription_plan as sp on sp.plan_id=s.plan_id
	JOIN product as p on p.product_id=sp.product_id
	JOIN invoice as i ON i.stripe_price_id=sp.stripe_price_id
	WHERE c.customer_id=$1 AND p.product_id=$2
	`


	var total int
	err := c.DB.QueryRow(query, cusId, productId).Scan(&total)
	if err != nil {
		return nil, err
	}

	return &total, nil
}

func (c *CustomerDB) GetSubInvoices(
	cusId int,
	productId int, 
	limit int,
) ([]models.Invoice, error) {
	query := fmt.Sprintf(`
	SELECT 
	i.invoice_id, i.paid, i.attempted, i.status, i.total, i.created, i.invoice_url, 
	s.sub_id, s.plan_id, s.start_date, 
	cc.card_id, cc.brand, cc.last4
	FROM customer as c 
	JOIN subscription as s on s.customer_id=c.customer_id
	JOIN subscription_plan as sp on sp.plan_id=s.plan_id
	JOIN product as p on p.product_id=sp.product_id
	JOIN invoice as i ON i.sub_id=s.sub_id
	JOIN customer_card as cc on cc.card_id=s.card_id
	WHERE c.customer_id=$1 AND p.product_id=$2
	ORDER BY i.created DESC
	LIMIT %d
	`, limit)


	rows, err := c.DB.Query(query, cusId, productId)
	if err != nil {
		return nil, err
	}

	invoices := []models.Invoice{}
	for rows.Next() {
		var in models.Invoice
		in.Subscription = &models.Subscription{}
		in.CardInfo = &models.CardInfo{}
		if err := rows.Scan(
			&in.ID, &in.Paid, &in.Attempted, &in.Status, &in.Total, &in.Created, &in.InvoiceURL,
			&in.Subscription.ID, &in.Subscription.PlanID, &in.Subscription.StartDate,
			&in.CardInfo.ID, &in.CardInfo.Brand, &in.CardInfo.Last4,
		); err != nil {
			return nil, err
		}
		invoices = append(invoices, in)
	}

	return invoices, nil
}

func (c *CustomerDB) GetSubscriptionUsages(
	cusId int,
	productId int, 
	limit int,
) ([]models.CusUsage, error) {
	query := fmt.Sprintf(`
	SELECT cu.usage_id, cu.created, su.title, p.product_id, p."name"
	FROM customer as c 
	JOIN subscription as s on s.customer_id=c.customer_id
	JOIN customer_usage as cu on cu.customer_uuid=c.uuid
	JOIN subscription_plan as sp on sp.plan_id=s.plan_id
	JOIN subscription_usage as su on su.plan_id=sp.plan_id
	JOIN product as p on p.product_id=sp.product_id
	WHERE c.customer_id=$1 AND p.product_id=$2
	ORDER BY cu.created DESC
	LIMIT %d
	`, limit)


	rows, err := c.DB.Query(query, cusId, productId)
	if err != nil {
		return nil, err
	}

	usages := []models.CusUsage{}
	for rows.Next() {
		var usage models.CusUsage
		if err := rows.Scan(
			&usage.ID, &usage.Created, 
			&usage.SubUsageTitle, &usage.ProductID, &usage.ProductName,
		); err != nil {
			return nil, err
		}
		usages = append(usages, usage)
	}

	return usages, nil
}

func (c *CustomerDB) AddNewCustomerCard(cusId int, cardInfo models.CardInfo) (*int, error) {
	query := `
	INSERT into customer_card (last4, stripe_id, customer_id, brand) VALUES ($1, $2, $3, $4) RETURNING card_id
	`
	
	var cardId int
	err := c.DB.QueryRow(query, 
		cardInfo.Last4,
		cardInfo.StripeID,
		cardInfo.CusID,
		cardInfo.Brand,
	).Scan(&cardId)
	if err != nil {
		return nil, err
	}

	return &cardId, nil
}

func (c *CustomerDB) UpdateCusDefaultCard(cusId int, cardId int) (error) {
	query := `
	UPDATE customer SET default_card_id=$1 WHERE customer_id=$2
	`
	
	_, err := c.DB.Exec(query, 
		cardId,
		cusId,
	)
	return err
}

func (c *CustomerDB) GetCusCard(cusId int, cardId int) (*models.CardInfo, error) {
	query := `
		SELECT last4, stripe_id, brand FROM customer_card WHERE card_id=$1 AND customer_id=$2
	`
	
	card := models.CardInfo{}
	err := c.DB.QueryRow(query, 
		cardId,
		cusId,
	).Scan(&card.Last4, &card.StripeID, &card.Brand)

	if err != nil {
		return nil, err
	}
	return &card, nil
}