package db

import (
	"database/sql"

	"github.com/johnyeocx/usual/server/db/models"
)

type BusinessDB struct {
	DB *sql.DB
}


func (businessDB *BusinessDB) GetBusinessByID(businessId int) (*models.Business, error) {

	selectStatement := `SELECT business_id, name, email, country, business_category, business_url, individual_id, stripe_id, description
	from business WHERE business_id=$1`

	var business models.Business
	if err := businessDB.DB.QueryRow(selectStatement, businessId).Scan(
		&business.BusinessID,
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

func (businessDB *BusinessDB) GetBusinessByEmail(email string) (*models.Business, error) {

	selectStatement := `SELECT 
	business_id, name, email, country, 
	business_category, business_url, individual_id, stripe_id, description
	from business WHERE email=$1`

	var business models.Business
	if err := businessDB.DB.QueryRow(selectStatement, email).Scan(
		&business.BusinessID,
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

func (b *BusinessDB) GetBusinessSubscribers(businessId int) (*[]models.Customer, error) {
	stmt := `SELECT DISTINCT c.customer_id, c.name, c.email FROM
	business as b 
	JOIN product as p on b.business_id=p.business_id
	JOIN subscription_plan as sp on sp.product_id=p.product_id
	JOIN subscription as s on sp.plan_id=s.plan_id
	JOIN customer as c on c.customer_id=s.customer_id
	WHERE b.business_id=$1`

	rows, err := b.DB.Query(stmt, businessId)
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
			&subscriber.Email,
		); err != nil {
			continue
        }
		
		subscribers = append(subscribers, subscriber)
    }

	
	return &subscribers, nil
}

func (businessDB *BusinessDB) CreateBusinessProfile(
	id int,
	businessCategory string, 
	businessUrl string, 
	individualId int,
	stripeId string,
) (error) {
	updateStatement := `UPDATE business SET business_category=$1, business_url=$2, individual_id=$3, stripe_id=$4
	WHERE business_id=$5 AND email_verified=true`

	_, err := businessDB.DB.Exec(updateStatement, 
		businessCategory, 
		businessUrl, 
		individualId, 
		stripeId,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (businessDB *BusinessDB) SetBusinessDescription(
	businessId int,
	description string,
) (error) {
	_, err := businessDB.DB.Exec(
		`UPDATE business SET description=$1 WHERE business_id=$2`, 
		description, 
		businessId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (b *BusinessDB) SetBusinessCategory(
	businessId int,
	category string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET business_category=$1 WHERE business_id=$2`, 
		category, businessId,
	)

	return err
}

func (b *BusinessDB) SetBusinessName(
	businessId int,
	name string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET name=$1 WHERE business_id=$2`, 
		name, businessId,
	)

	return err
}

func (b *BusinessDB) SetBusinessEmail(
	businessId int,
	email string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET email=$1 WHERE business_id=$2`, 
		email, businessId,
	)

	return err
}

func (b *BusinessDB) SetBusinessCountry(
	businessId int,
	countryCode string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET country=$1 WHERE business_id=$2`, 
		countryCode, businessId,
	)

	return err
}

func (b *BusinessDB) SetBusinessUrl(
	businessId int,
	url string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET business_url=$1 WHERE business_id=$2`, 
		url, businessId,
	)

	return err
}