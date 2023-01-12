package customer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
)


func CreateCFromSubscribe(
	sqlDB *sql.DB,
	name string,
	email string,
	card *models.CreditCard,
) (*int, *models.RequestError) {

	// 1. check if customer already created
	c := db.CustomerDB{DB: sqlDB}
	err := c.GetCustomerByEmail(email) 

	if err == nil || err != sql.ErrNoRows {
		return  nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusConflict,
		}
	}
	
	newC := models.Customer {
		Name: name,
		Email: email,
	}
	// 2. Create stripe customer
	stripeId, err := my_stripe.CreateCustomer(&newC, card)
	if err != nil {
		errMap := map[string]interface{}{}
		json.Unmarshal([]byte(err.Error()), &errMap)
		fmt.Println(errMap["code"])

		// handle card declined

		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 3. Insert into db
	cId, err := c.CreateCFromSubscribe(name, email, *stripeId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	} else {
		return cId, nil
	}
}