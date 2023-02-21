package cusdb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/johnyeocx/usual/server/db/models"
)


type CustomerDB struct {
	DB	*sql.DB
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
	c.customer_id, c.first_name, c.last_name, c.email, c.stripe_id, c.default_card_id, 
	c.address_line1, c.address_line2, c.postal_code, c.city, c.country, c.uuid, c.signin_provider
	FROM customer as c 
	WHERE customer_id=$1
	GROUP BY c.customer_id`, 
	cusId).Scan(
		&cus.ID,
		&cus.FirstName,
		&cus.LastName,
		&cus.Email,
		&cus.StripeID,
		&cus.DefaultCardID,
		&cus.Address.Line1,
		&cus.Address.Line2,
		&cus.Address.PostalCode,
		&cus.Address.City,
		&cus.Address.Country,
		&cus.Uuid,
		&cus.SignInProvider,
	)
	if err != nil {
		return nil, err
	}

	return &cus, nil
}

func (c *CustomerDB) GetCustomerByEmail (
	email string,
) (*models.Customer, error) {
	var cus models.Customer 
	err := c.DB.QueryRow(`SELECT 
		customer_id, first_name, last_name, email, uuid, email_verified, signin_provider
		FROM customer WHERE email=$1`, 
	email).Scan(
		&cus.ID,
		&cus.FirstName,
		&cus.LastName,
		&cus.Email,
		&cus.Uuid,
		&cus.EmailVerified,
		&cus.SignInProvider,
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
) (*models.Customer, error) {

	var cus models.Customer
	err := c.DB.QueryRow("SELECT customer_id, password, signin_provider FROM customer WHERE email=$1", 
		email,
	).Scan(&cus.ID, &cus.Password, &cus.SignInProvider)

	if err != nil {
		return nil, err
	}
	return &cus, nil
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

func (c *CustomerDB) CusOwnsCard (
	customerId int,
	cardId int,
) (*models.Customer, *string, error) {
	var cus models.Customer
	var cardStripeId string
	err := c.DB.QueryRow(`SELECT c.stripe_id, cc.stripe_id, c.default_card_id  FROM 
		customer as c JOIN customer_card as cc ON c.customer_id=cc.customer_id
		WHERE c.customer_id=$1 AND cc.card_id=$2`, 
		customerId, cardId,
	).Scan(&cus.StripeID, &cardStripeId, &cus.DefaultCardID)

	if err != nil {
		return nil, nil, err
	}
	return &cus, &cardStripeId, nil
}

func (c *CustomerDB) CardIsBeingUsed (
	cardId int,
) (*string, error) {
	var cardStripeId string
	err := c.DB.QueryRow(`SELECT cc.stripe_id FROM 
		customer_card as cc JOIN subscription as s on s.card_id=cc.card_id
		WHERE cc.card_id=$1 AND s.cancelled = 'FALSE'`, cardId,
	).Scan(&cardStripeId)

	if err != nil {
		return nil, err
	}

	return &cardStripeId, nil
}



func (c *CustomerDB) GetCustomerSubscriptions(cusId int) ([]models.Subscription, error) {
	query := `
	WITH ranked_table AS (

		SELECT 
		c.customer_id,
		s.sub_id, s.start_date, s.cancelled, s.expires, s.cancelled_date, s.card_id,
		b.name, b.business_id,
		p.product_id, p.name, p.description, p.category_id, pc.title,
		sp.plan_id, sp.recurring_interval, sp.recurring_interval_count, sp.unit_amount, sp.currency,
		i.invoice_id, i.created, i.status, i.total, i.invoice_url, i.card_id, i.payment_intent_status,
		ROW_NUMBER() OVER 
		(PARTITION BY s.sub_id ORDER BY i.created DESC) as rank
		FROM customer as c 
		JOIN subscription as s ON c.customer_id=s.customer_id
		JOIN invoice as i ON i.sub_id=s.sub_id
		JOIN subscription_plan as sp ON sp.plan_id=s.plan_id
		JOIN product as p ON p.product_id=sp.product_id
		JOIN product_category as pc ON pc.category_id=p.category_id
		JOIN business as b ON b.business_id=p.business_id
		GROUP BY sp.plan_id, p.product_id, s.sub_id, i.invoice_id, i.stripe_in_id, c.customer_id, pc.category_id, b.business_id
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
			&product.ProductID, &product.Name, &product.Description, &product.CategoryID, &product.CatTitle,
			&plan.PlanID, &plan.RecurringDuration.Interval, &plan.RecurringDuration.IntervalCount, &plan.UnitAmount, &plan.Currency,
			&invoice.ID, &invoice.Created, &invoice.Status, &invoice.Total, &invoice.InvoiceURL, &invoice.CardID, &invoice.PaymentIntentStatus,
			&rank,
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
		cc.card_id, cc.last4, cc.brand, cc.deleted FROM customer as c JOIN customer_card as cc on c.customer_id=cc.customer_id
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
			&card.Deleted,
			
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
	i.sub_id, i.card_id, i.payment_intent_status
	from invoice as i
	JOIN customer as c ON i.stripe_cus_id=c.stripe_id
	
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
		// var product models.Product
		in.CardInfo = &models.CardInfo{}
		if err := rows.Scan(
			&in.ID, &in.Paid, &in.Attempted, &in.Status, &in.Total, &in.Created, &in.InvoiceURL,
			&in.SubID,
			&in.CardID, &in.PaymentIntentStatus,
		); err != nil {
			return nil, err
		}
		// in.Subscription.SubProduct = &models.SubscriptionProduct{
		// 	Product: product,
		// }
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
	JOIN invoice as i ON i.sub_id=s.sub_id
	WHERE c.customer_id=$1 AND p.product_id=$2 AND i.payment_intent_status='succeeded'
	`

	var total sql.NullInt64
	err := c.DB.QueryRow(query, cusId, productId).Scan(&total)
	if err != nil {
		return nil, err
	}

	var totalInt int = 0
	if total.Valid {
		totalInt = int(total.Int64)
	}
	return &totalInt, nil
}

func (c *CustomerDB) CusHasPaidSubBefore(
	cusId int,
	subId int, 
) ([]models.Invoice, error) {
	query := `
	SELECT 
	i.invoice_id, i.paid, i.attempted, i.status, i.total, i.created, i.invoice_url, 
	i.sub_id, i.card_id
	FROM customer as c 
	JOIN subscription as s on s.customer_id=c.customer_id
	JOIN invoice as i ON i.sub_id=s.sub_id
	WHERE c.customer_id=$1 AND i.sub_id=$2 AND i.status='paid'
	ORDER BY i.created DESC
	`

	rows, err := c.DB.Query(query, cusId, subId)
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
			&in.SubID,
			&in.CardID,
		); err != nil {
			return nil, err
		}
		invoices = append(invoices, in)
	}

	return invoices, nil
}

func (c *CustomerDB) GetSubInvoices(
	cusId int,
	productId int, 
	limit int,
) ([]models.Invoice, error) {
	query := fmt.Sprintf(`
	SELECT 
	i.invoice_id, i.paid, i.attempted, i.status, i.total, i.created, i.invoice_url, 
	i.sub_id, i.card_id, i.payment_intent_status
	FROM customer as c 
	JOIN subscription as s on s.customer_id=c.customer_id
	JOIN subscription_plan as sp on sp.plan_id=s.plan_id
	JOIN product as p on p.product_id=sp.product_id
	JOIN invoice as i ON i.sub_id=s.sub_id
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
			&in.SubID,
			&in.CardID, &in.PaymentIntentStatus,
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

func (c *CustomerDB) GetCusFCMToken(cusId int) (*string, error) {
	query := `
		SELECT token FROM customer_fcm_token WHERE customer_id=$1
	`
	
	var token string
	err := c.DB.QueryRow(query, 
		cusId,
	).Scan(&token)

	if err != nil {
		return nil, err
	}
	return &token, nil
}