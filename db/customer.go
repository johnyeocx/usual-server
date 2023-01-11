package db

import "database/sql"

type CustomerDB struct {
	DB	*sql.DB
}

func (c *CustomerDB) GetCustomerByEmail (
	email string,
) (error) {
	var res string 
	err := c.DB.QueryRow("SELECT email FROM customer WHERE email=$1", 
	email).Scan(&res)

	return err
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