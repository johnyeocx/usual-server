package db

import (
	"database/sql"
	"errors"
	"fmt"

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
) (*int, error) {
	hashedPassword, err := secure.GenerateHashFromStr(password)
	if err != nil {
		return nil, err
	}

	var cusId int
	err = c.DB.QueryRow(`
		INSERT into customer (name, email, password) VALUES ($1, $2, $3) RETURNING customer_id`,
		name, email, hashedPassword,
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
	err := c.DB.QueryRow(`SELECT 
		customer_id, name, email, stripe_id, default_card_id
		FROM customer WHERE customer_id=$1`, 
	cusId).Scan(
		&cus.ID,
		&cus.Name,
		&cus.Email,
		&cus.StripeID,
		&cus.DefaultCardID,
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
		customer_id, name, email
		FROM customer WHERE email=$1`, 
	email).Scan(
		&cus.ID,
		&cus.Name,
		&cus.Email,
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

func (c *CustomerDB) GetCustomerHashedPassword (
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
	query := `SELECT 
	s.sub_id, s.start_date, 
	b.name, b.business_id, 
	p.product_id, p.name, p.description,
	sp.plan_id, sp.product_id,
	sp.recurring_interval, sp.recurring_interval_count, sp.unit_amount, sp.currency
	from customer as c
	JOIN subscription as s on c.customer_id=s.customer_id
	JOIN subscription_plan as sp on s.plan_id=sp.plan_id
	JOIN product as p on sp.product_id=p.product_id
	JOIN business as b on b.business_id=p.business_id
	WHERE c.customer_id=$1`
	
	
	rows, err := c.DB.Query(query, cusId)
	if err != nil {
		return nil, err
	}

	subs := []models.Subscription{}
	for rows.Next() {
		var sub models.Subscription
		sub.SubProduct = &models.SubscriptionProduct{}
		if err := rows.Scan(
			&sub.ID,
			&sub.StartDate,
			&sub.BusinessName, &sub.BusinessID,
			&sub.SubProduct.Product.ProductID, &sub.SubProduct.Product.Name, &sub.SubProduct.Product.Description,
			&sub.SubProduct.SubPlan.PlanID, &sub.SubProduct.SubPlan.ProductID,
			&sub.SubProduct.SubPlan.RecurringDuration.Interval,
			&sub.SubProduct.SubPlan.RecurringDuration.IntervalCount,
			&sub.SubProduct.SubPlan.UnitAmount,
			&sub.SubProduct.SubPlan.Currency,
		); err != nil {
			return nil, err
		}
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
	cc.card_id, cc.brand, cc.last4
	from invoice as i
	JOIN customer as c ON i.stripe_cus_id=c.stripe_id
	JOIN subscription as s on s.stripe_sub_id=i.stripe_sub_id
	JOIN customer_card as cc on cc.card_id=s.card_id
	WHERE c.customer_id=$1 
	ORDER BY created ASC
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