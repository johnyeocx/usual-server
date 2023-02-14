package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/johnyeocx/usual/server/db/models"
)

type SubscriptionDB struct {
	DB *sql.DB
}





func (s *SubscriptionDB) GetCreateSubData(
	productId int,
)(*models.SubscriptionProduct, *string, error) {
	selectStatement := `SELECT 
	p.product_id, p.business_id, p.name, p.description, p.category_id, p.stripe_product_id,
	sp.plan_id, sp.currency, sp.recurring_interval, sp.recurring_interval_count, sp.unit_amount, sp.stripe_price_id, b.stripe_id
	FROM product as p
	JOIN subscription_plan as sp on p.product_id = sp.product_id
	JOIN business as b on b.business_id=p.business_id
	WHERE p.product_id=$1`

	var product models.Product
	var subPlan models.SubscriptionPlan
	var stripeBusId string
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
		&stripeBusId,
	)

	if err != nil {
		return nil, nil, err
	}
	
	subProduct := models.SubscriptionProduct{
		Product: product,
		SubPlan: subPlan,
	}
	return &subProduct, &stripeBusId, nil
}


func (s *SubscriptionDB) GetCusResumeSubData(cusId int, subId int) (
	map[string]interface{},
	error,
) {
	query := `SELECT 
	c.stripe_id, b.stripe_id, cc.stripe_id, 

	s.stripe_sub_id, s.start_date, s.cancelled, s.card_id, s.expires,
	sp.recurring_interval, sp.recurring_interval_count, sp.stripe_price_id, i.created

	from customer as c
	JOIN subscription as s on c.customer_id=s.customer_id
	JOIN subscription_plan as sp on sp.plan_id=s.plan_id
	JOIN product as p on p.product_id=sp.product_id
	JOIN business as b on b.business_id=p.business_id
	JOIN invoice as i on i.stripe_prod_id=p.stripe_product_id
	JOIN customer_card as cc on cc.customer_id=c.customer_id
	WHERE c.customer_id=$1 AND s.sub_id=$2
	ORDER BY i.created DESC 
	LIMIT 1`
	
	var sub models.Subscription
	subPlan := models.SubscriptionPlan{}
	subPlan.RecurringDuration = models.TimeFrame{}
	invoice := models.Invoice{}
	var cusStripeId string
	var busStripeId string
	var cardStripeId string

	err := s.DB.QueryRow(query, cusId, subId).Scan(
		&cusStripeId,
		&busStripeId,
		&cardStripeId,
		&sub.StripeSubID,
		&sub.StartDate,
		&sub.Cancelled,
		&sub.CardID,
		&sub.Expires,
		&subPlan.RecurringDuration.Interval,
		&subPlan.RecurringDuration.IntervalCount,
		&subPlan.StripePriceID,
		&invoice.Created,
	)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"stripe_cus_id": cusStripeId,
		"stripe_bus_id": busStripeId,
		"stripe_card_id": cardStripeId,
		"subscription": sub,
		"plan": subPlan,
		"last_invoice": invoice,
	}, nil
}


func (s *SubscriptionDB) CusOwnsSub(cusId int, subId int) (
	*models.Subscription, 
	*models.SubscriptionPlan, 
	*models.Invoice,
	error,
) {

	query := `SELECT s.stripe_sub_id, s.start_date, s.cancelled, s.card_id,
	sp.recurring_interval, sp.recurring_interval_count, i.created
	from customer as c
	JOIN subscription as s on c.customer_id=s.customer_id
	JOIN subscription_plan as sp on sp.plan_id=s.plan_id
	JOIN product as p on sp.product_id=p.product_id
	LEFT JOIN invoice as i on i.sub_id=s.sub_id
	WHERE c.customer_id=$1 AND s.sub_id=$2
	ORDER BY i.created DESC 
	LIMIT 1`
	
	var sub models.Subscription
	subPlan := models.SubscriptionPlan{}
	subPlan.RecurringDuration = models.TimeFrame{}
	
	invoiceCreated := sql.NullTime{}
	err := s.DB.QueryRow(query, cusId, subId).Scan(
		&sub.StripeSubID,
		&sub.StartDate,
		&sub.Cancelled,
		&sub.CardID,
		&subPlan.RecurringDuration.Interval,
		&subPlan.RecurringDuration.IntervalCount,
		&invoiceCreated,
	)

	var invoice models.Invoice
	if invoiceCreated.Valid {
		invoice.Created = invoiceCreated.Time
	}
	
	if err != nil {
		return nil, nil, nil, err
	}
	return &sub, &subPlan, &invoice, nil
}

func (s *SubscriptionDB) InsertSubscriptions(subs *[]models.Subscription) (
	[]models.Subscription, error,
) {

	numCols := 5

	valueStrings := make([]string, 0, len(*subs))
    valueArgs := make([]interface{}, 0, len(*subs) * numCols)
	
    for i, sub := range (*subs) {
		j := i * numCols + 1
		valueString := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", j, j + 1, j + 2, j + 3, j + 4)
        valueStrings = append(valueStrings, valueString)
        valueArgs = append(valueArgs, sub.StripeSubID)
        valueArgs = append(valueArgs, sub.CustomerID)
        valueArgs = append(valueArgs, sub.PlanID)
        valueArgs = append(valueArgs, sub.StartDate)
        valueArgs = append(valueArgs, sub.CardID)
    }
	
    stmt := fmt.Sprintf(
		`INSERT into subscription (stripe_sub_id, customer_id, plan_id, start_date, card_id) VALUES %s RETURNING sub_id`, 
		strings.Join(valueStrings, ","))
	
	returnedSubs := []models.Subscription{}
    rows, err := s.DB.Query(stmt, valueArgs...)
	if err != nil {
		return nil, err
	}

	index := 0
	for rows.Next() {
		var subId int
		if err := rows.Scan(&subId); err != nil {
			// TO FIX
			continue
		}
		sub := (*subs)[index]
		sub.ID = subId
		returnedSubs = append(returnedSubs, sub)
	}

	return returnedSubs, nil
}

func (s *SubscriptionDB) CancelSubscription(subId int, expires time.Time) (error) {

	stmt := `
		UPDATE subscription SET cancelled=$1, expires=$2, cancelled_date=$3 WHERE sub_id=$4
	`
	_, err := s.DB.Exec(stmt, true, expires, time.Now(), subId)
	return err
}

func (s *SubscriptionDB) ResumeSubscription(subId int, cardId int, stripeSubId string) (error) {

	stmt := `
		UPDATE subscription SET cancelled='FALSE', expires=NULL, cancelled_date=NULL, card_id=$1, stripe_sub_id=$2 WHERE sub_id=$3
	`
	_, err := s.DB.Exec(stmt, cardId, stripeSubId, subId)
	return err
}

func (s *SubscriptionDB) DeleteSubscriptionAndInvoices(subId int) (error) {
	stmt := `DELETE FROM subscription WHERE sub_id=$1`
	_, err := s.DB.Exec(stmt, subId)

	if err != nil {
		return err
	}

	stmt2 := `DELETE FROM invoice WHERE sub_id=$1`
	_, err = s.DB.Exec(stmt2, subId)
	return err
}

func (s *SubscriptionDB) UpdateSubCardID(subId int, cardId int) (error) {
	stmt := `UPDATE subscription SET card_id=$1 WHERE sub_id=$2`
	_, err := s.DB.Exec(stmt, cardId, subId)
	return err
}

func (s *SubscriptionDB) GetSubInvoicesFromSubID(
	subId int, 
	limit int,
) ([]models.Invoice, error) {
	query := fmt.Sprintf(`
	SELECT 
	i.invoice_id, i.paid, i.attempted, i.status, i.total, i.created, i.invoice_url, 
	i.sub_id, i.card_id, i.payment_intent_status
	FROM invoice as i WHERE i.sub_id=$1 LIMIT %d
	`, limit)


	rows, err := s.DB.Query(query, subId)
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