package busdb

import (
	"database/sql"
	"fmt"

	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/db/models/bus_models"
)

// GET BUSINESS FROM PARAM

func (b *BusinessDB) GetBusinessAndTotal(bId int) (*models.Business, *models.JsonNullInt64, *models.JsonNullInt64, error) {

	selectStatement := `
	WITH table1 as (
		SELECT SUM(bp.amount) as payout_total,
		b.business_id, b.name, b.email, b.country, b.business_category, b.business_url, b.individual_id, b.stripe_id, b.description,
		b.external_account_id, b.external_account_type
		FROM business as b 
		LEFT JOIN business_payout as bp on bp.business_id=b.business_id
		GROUP BY b.business_id
	) 
	SELECT 
	b.payout_total,  SUM(i.total) as received_total,
	b.business_id, b.name, b.email, b.country, b.business_category, b.business_url, b.individual_id, b.stripe_id, b.description,
	b.external_account_id, b.external_account_type
	from table1 as b
	LEFT JOIN product as p on p.business_id=b.business_id
	LEFT JOIN invoice as i on i.stripe_prod_id=p.stripe_product_id
	WHERE b.business_id=$1
	GROUP BY b.payout_total,
	b.business_id, b.name, b.email, b.country, b.business_category, b.business_url, b.individual_id, b.stripe_id, b.description,
	b.external_account_id, b.external_account_type
	`

	var business models.Business
	var payoutTotal models.JsonNullInt64
	var receivedTotal models.JsonNullInt64

	if err := b.DB.QueryRow(selectStatement, bId).Scan(
		&payoutTotal,
		&receivedTotal,
		&business.ID,
		&business.Name,
		&business.Email,
		&business.Country,
		&business.BusinessCategory,
		&business.BusinessUrl,
		&business.IndividualID,
		&business.StripeID,
		&business.Description,
		&business.ExternalAccountID,
		&business.ExternalAccountType,
	); err != nil {
		return nil, nil, nil, err
	}

	if (!payoutTotal.Valid) {
		payoutTotal.Valid = true
		payoutTotal.Int64 = 0
	}

	if (!receivedTotal.Valid) {
		receivedTotal.Valid = true
		receivedTotal.Int64 = 0
	}

	return &business, &payoutTotal, &receivedTotal, nil
}

func (b *BusinessDB) GetBusinessByID(businessId int) (*models.Business, error) {

	selectStatement := `SELECT 
	business_id, name, email, country, business_category, business_url, individual_id, stripe_id, description
	from business WHERE business_id=$1`

	var business models.Business
	if err := b.DB.QueryRow(selectStatement, businessId).Scan(
		&business.ID,
		&business.Name,
		&business.Email,
		&business.Country,
		&business.BusinessCategory,
		&business.BusinessUrl,
		&business.IndividualID,
		&business.StripeID,
		&business.Description,
	); err != nil {
		return nil, err
	}

	return &business, nil
}

func (b *BusinessDB) GetBusinessEmailVerified(email string) (bool, error) {
	
	var verified bool
	err := b.DB.QueryRow("SELECT email_verified FROM business WHERE email=$1", 
	email).Scan(&verified)

	return verified, err
}

func (b *BusinessDB) GetBusinessByName(name string) (*models.Business, error) {

	selectStatement := `SELECT 
	business_id, name, email, country, 
	business_category, business_url, individual_id, stripe_id, description, email_verified
	from business WHERE name=$1`

	var business models.Business
	if err := b.DB.QueryRow(selectStatement, name).Scan(
		&business.ID,
		&business.Name,
		&business.Email,
		&business.Country,
		&business.BusinessCategory,
		&business.BusinessUrl,
		&business.IndividualID,
		&business.StripeID,
		&business.Description,
		&business.EmailVerified,
	); err != nil {
		return nil, err
	}

	return &business, nil
}

func (b *BusinessDB) GetBusinessByEmail(email string) (*models.Business, error) {

	selectStatement := `SELECT 
	business_id, name, email, country, 
	business_category, business_url, individual_id, stripe_id, description, email_verified
	from business WHERE email=$1`

	var business models.Business
	if err := b.DB.QueryRow(selectStatement, email).Scan(
		&business.ID,
		&business.Name,
		&business.Email,
		&business.Country,
		&business.BusinessCategory,
		&business.BusinessUrl,
		&business.IndividualID,
		&business.StripeID,
		&business.Description,
		&business.EmailVerified,
	); err != nil {
		return nil, err
	}

	return &business, nil
}

func (b *BusinessDB) GetBusinessFromBankStripeID(bankAccountId string) (*models.Business, *int, error) {

	selectStatement := `SELECT 
	ba.bank_account_id, b.business_id, b.name, b.email, b.country, b.business_category, b.business_url, b.individual_id, b.stripe_id, b.description
	from business as b JOIN business_bank_account as ba ON b.business_id=ba.business_id 
	WHERE ba.stripe_id=$1
	`

	var business models.Business
	var baID int
	if err := b.DB.QueryRow(selectStatement, bankAccountId).Scan(
		&baID,
		&business.ID,
		&business.Name,
		&business.Email,
		&business.Country,
		&business.BusinessCategory,
		&business.BusinessUrl,
		&business.IndividualID,
		&business.StripeID,
		&business.Description,
	); err != nil {
		return nil, nil, err
	}

	return &business, &baID, nil
}

// GET BUSINESS DETAILS

func (businessDB *BusinessDB) GetBusinessPasswordByID(businessId int) (*string, error) {

	selectStatement := `SELECT password
	from business WHERE business_id=$1`

	var password string
	if err := businessDB.DB.QueryRow(selectStatement, businessId).Scan(
		&password,
	); err != nil {
		return nil, err
	}

	return &password, nil
}

func (b *BusinessDB) GetBusinessStripeID(businessId int) (*string, error) {


	var stripeId string
	err := b.DB.QueryRow(`SELECT stripe_id FROM business WHERE business_id=$1`, 
	businessId).Scan(&stripeId)

	if err != nil {
		return nil, err
	} else {
		return &stripeId, nil
	}
}

// GET BUSINESS STATS

func (b *BusinessDB) GetBusinessSubs(businessId int) (*[]models.SubInfo, error) {
	stmt := `
	SELECT c.customer_id, c.first_name, c.last_name, c.email, p.product_id, p."name", 
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
			&info.Email,
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

func (b *BusinessDB) GetBusinessBankAccounts(businessId int) ([]models.BankAccount, error) {
	stmt := `SELECT 
		ba.bank_account_id, ba.account_holder_name, ba.bank_name, ba.last4, ba.routing_number
		FROM
		business_bank_account as ba 
		JOIN business as b on b.business_id=ba.business_id
		WHERE b.business_id=$1`

	rows, err := b.DB.Query(stmt, businessId)
	if err != nil {
		return nil, err
	}
	

	defer rows.Close()

	bankAccounts := []models.BankAccount{}
	for rows.Next() {
		ba := models.BankAccount{}

        if err := rows.Scan(
			&ba.ID,
			&ba.AccountHolderName,
			&ba.BankName,
			&ba.Last4,
			&ba.RoutingNumber,
		); err != nil {
			fmt.Println(err)
			continue
        }
		
		bankAccounts = append(bankAccounts, ba)
    }

	
	return bankAccounts, nil
}

func (b *BusinessDB) GetBusinessInvoices(businessId int, limit int) ([]models.InvoiceData, error) {
	stmt := fmt.Sprintf(`SELECT 
	c.customer_id, c.first_name, c.last_name, 
	i.invoice_id, i.total, i.invoice_url, i.status, i.attempted, i.app_fee_amt,
	p.name, p.product_id, i.payment_intent_status, i.created
	FROM business as b
	JOIN product as p ON b.business_id=p.business_id
	JOIN subscription_plan as sp on sp.product_id=p.product_id
	JOIN invoice as i on i.stripe_prod_id=p.stripe_product_id
	JOIN customer as c on i.stripe_cus_id=c.stripe_id
	WHERE b.business_id=$1
	ORDER BY i.created DESC
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
			&invoice.PaymentIntentStatus,
			&invoice.Created,
		); err != nil {
			fmt.Println(err)
			continue
        }
		
		invoices = append(invoices, invoice)
    }
	
	return invoices, nil
}

func (b *BusinessDB) GetBusinessPayouts(businessId int, limit int) ([]bus_models.BusinessPayout, error) {
	stmt := fmt.Sprintf(`SELECT 
	bp.payout_id, bp.amount, bp.currency, bp.status, bp.arrival_date, bp.stripe_dest_id, bp.type, bp.external_account_id
	FROM business as b
	JOIN business_payout as bp on b.business_id=bp.business_id
	WHERE b.business_id=$1
	LIMIT %d`, limit)

	rows, err := b.DB.Query(stmt, businessId)
	if err != nil {
		return nil, err
	}
	

	defer rows.Close()

	payouts := []bus_models.BusinessPayout{}
	for rows.Next() {
		var p bus_models.BusinessPayout
		
        if err := rows.Scan(
			&p.ID,
			&p.Amount,
			&p.Currency,
			&p.Status,
			&p.ArrivalDate,
			&p.StripeDestID,
			&p.Type,
			&p.ExternalAccountID,
		); err != nil {
			fmt.Println(err)
			continue
        }
		
		payouts = append(payouts, p)
    }
	
	return payouts, nil
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