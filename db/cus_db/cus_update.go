package cusdb

import (
	"fmt"
	"time"

	"github.com/johnyeocx/usual/server/db/models"
)

func (c *CustomerDB) UpdateCusDefaultCard(
	cusId int, 
	cardId int,
) (error) {
	query := `
	UPDATE customer SET default_card_id=$1 WHERE customer_id=$2
	`
	
	_, err := c.DB.Exec(query, 
		cardId,
		cusId,
	)
	return err
}

func (c *CustomerDB) SetCardDeleted(
	cardId int,
) (error) {
	query := `
	UPDATE customer_card SET deleted=$1 WHERE card_id=$2
	`
	
	_, err := c.DB.Exec(query, 
		true,
		cardId,
	)
	return err
}

func (c *CustomerDB) UpdateCusName(
	cusId int, 
	firstName string, 
	lastName string,
) (error){

	_, err := c.DB.Exec(`UPDATE customer SET first_name=$1, last_name=$2 WHERE customer_id=$3`, firstName, lastName, cusId)
	return err
}

func (c *CustomerDB) UpdateCusEmail(cusId int, email string) (error){
	_, err := c.DB.Exec(`UPDATE customer SET email=$1 WHERE customer_id=$2`, email, cusId)
	return err
}

func (c *CustomerDB) UpdateCusAddress(cusId int, address models.Address) (error){
	fmt.Println(address)
	_, err := c.DB.Exec(`UPDATE customer 
			SET address_line1=$1, address_line2=$2, postal_code=$3, city=$4, country=$5 WHERE
		 customer_id=$6`, address.Line1, address.Line2, address.PostalCode, address.City, address.Country, cusId)
	return err
}

func (c *CustomerDB) UpdateCusCountry(cusId int, country string) (error){
	_, err := c.DB.Exec(`UPDATE customer 
	SET country=$1
	customer_id=$2`, country, cusId)
	return err
}

func (c *CustomerDB) UpdateCusPassword(cusId int, passwordHash string) (error){
	_, err := c.DB.Exec(`UPDATE customer 
	SET password=$1
	WHERE customer_id=$2 AND signin_provider='custom'`, passwordHash, cusId)
	return err
}

func (c *CustomerDB) InsertOrUpdateCusFCMToken(cusId int, fcmToken string) (error) {
	_, err := c.DB.Exec(`INSERT into customer_fcm_token 
	(token, customer_id, last_updated) VALUES($1, $2, $3)
	ON CONFLICT (customer_id) DO UPDATE SET token=$1, last_updated=$3`, fcmToken, cusId, time.Now())
	return err
}