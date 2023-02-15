package db

import (
	"database/sql"
	"fmt"

	"github.com/johnyeocx/usual/server/db/models"
)

func (b *BusinessDB) GetBusinessSubs(businessId int) (*[]models.SubInfo, error) {
	stmt := `
	SELECT c.customer_id, c.first_name, c.last_name, p.product_id, p."name", 
	s.start_date, s.sub_id
	FROM
	business as b 
	JOIN product as p on b.business_id=p.business_id
	JOIN subscription_plan as sp on sp.product_id=p.product_id
	JOIN subscription as s on sp.plan_id=s.plan_id
	JOIN customer as c on c.customer_id=s.customer_id
	WHERE b.business_id=$1 ORDER BY s.start_date DESC
	`
		
	rows, err := b.DB.Query(stmt, businessId)
	if err != nil {
		return nil, err
	}
	
	defer rows.Close()

	subInfos := []models.SubInfo{}
	for rows.Next() {
		var info models.SubInfo
		info.Subscription = models.Subscription{}
        if err := rows.Scan(
			&info.Subscription.CustomerID,
			&info.FirstName,
			&info.LastName,
			&info.ProductID,
			&info.ProductName,
			&info.Subscription.StartDate,
			&info.Subscription.ID,
		); err != nil {
			continue
        }
		
		subInfos = append(subInfos, info)
    }
	
	return &subInfos, nil
}

func (b *BusinessDB) GetBusinessInvoices(businessId int, limit int) ([]models.InvoiceData, error) {
	stmt := fmt.Sprintf(`SELECT 
	c.customer_id, c.first_name, c.last_name, 
	i.invoice_id, i.total, i.invoice_url, i.status, i.attempted, i.app_fee_amt,
	p.name, p.product_id
	FROM business as b
	JOIN product as p ON b.business_id=p.business_id
	JOIN subscription_plan as sp on sp.product_id=p.product_id
	JOIN invoice as i on i.stripe_prod_id=p.stripe_product_id
	JOIN customer as c on i.stripe_cus_id=c.stripe_id
	WHERE b.business_id=$1
	LIMIT %d`, limit)

	rows, err := b.DB.Query(stmt, businessId)
	if err != nil {
		return nil, err
	}
	

	defer rows.Close()

	invoices := []models.InvoiceData{}
	for rows.Next() {
		var invoice models.InvoiceData
		
        if err := rows.Scan(
			&invoice.CustomerID,
			&invoice.CusFirstName,
			&invoice.CusLastName,
			&invoice.InvoiceID,
			&invoice.Total,
			&invoice.InvoiceURL,
			&invoice.Status,
			&invoice.Attempted,
			&invoice.ApplicationFeeAmt,
			&invoice.ProductName,
			&invoice.ProductID,
		); err != nil {
			fmt.Println(err)
			continue
        }
		
		invoices = append(invoices, invoice)
    }
	
	return invoices, nil
}

func (b *BusinessDB) GetBusinessUsages(businessId int, limit int) ([]models.UsageInfo, error) {
	stmt := fmt.Sprintf(`SELECT 
		c.uuid, c.first_name, c.last_name, 
		cu.created, 
		su.title, su.sub_usage_id, su.unlimited, su.interval, su.amount, 
		p.product_id, p.name 
		FROM
		business as b
		JOIN product as p ON b.business_id=p.business_id
		JOIN subscription_plan as sp on sp.product_id=p.product_id
		JOIN subscription_usage as su on su.plan_id=sp.plan_id
		JOIN customer_usage as cu on cu.sub_usage_id=su.sub_usage_id
		JOIN customer as c on c.uuid=cu.customer_uuid
		WHERE b.business_id=$1
		LIMIT %d  
	`, limit)

	rows, err := b.DB.Query(stmt, businessId)
	if err != nil {
		return nil, err
	}
	

	defer rows.Close()

	usageInfos := []models.UsageInfo{}
	for rows.Next() {
		u := models.UsageInfo{}

        if err := rows.Scan(
			&u.CusUUID,
			&u.CusFirstName,
			&u.CusLastName,
			&u.Created,

			&u.SubUsage.Title,
			&u.SubUsage.ID,
			&u.SubUsage.Unlimited,
			&u.SubUsage.Interval,
			&u.SubUsage.Amount,
			&u.ProductID,
			&u.ProductName,
		); err != nil {
			fmt.Println(err)
			continue
        }
		
		usageInfos = append(usageInfos, u)
    }
	
	return usageInfos, nil
}

func (b *BusinessDB) GetBusinessTotalReceived(businessId int) (*int, error) {
	stmt := `
	SELECT SUM(total) from 
	business as b
	JOIN product as p on p.business_id=b.business_id
	JOIN invoice as i on i.stripe_prod_id=p.stripe_product_id
	WHERE b.business_id=$1
	`

	var total sql.NullInt64
	err := b.DB.QueryRow(stmt, businessId).Scan(&total)

	if err != nil {
		return nil, err
	} else if !total.Valid {
		zero := 0
		return &zero, nil
	} else {
		intTotal := int(total.Int64)
		return &intTotal, nil
	}
}