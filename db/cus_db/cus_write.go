package cusdb

import (
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/secure"
)


func (c *CustomerDB) CreateCustomer (
	firstName string,
	lastName string,
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
		INSERT into customer (first_name, last_name, email, password, uuid) VALUES ($1, $2, $3, $4, $5) RETURNING customer_id`,
		firstName, lastName, email, hashedPassword, uuid,
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
