package db

import (
	"database/sql"

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
		customer_id, name, email, stripe_id
		FROM customer WHERE customer_id=$1`, 
	cusId).Scan(
		&cus.ID,
		&cus.Name,
		&cus.Email,
		&cus.StripeID,
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


