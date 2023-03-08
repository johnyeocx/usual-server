package busdb

import (
	"time"

	"github.com/johnyeocx/usual/server/db/models"
)

func (businessDB *BusinessDB) CreateIndividual(individual *models.Person) (*int, error) {

	insertStatement := `INSERT into individual 
	(first_name, last_name, dialing_code, mobile_number, dob, 
		address_line1, address_line2, postal_code, city) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING individual_id`

	dob := time.Date(
		individual.DOB.Year, time.Month(individual.DOB.Month), individual.DOB.Day, 
		0, 0, 0, 0, time.UTC,
	)

	var insertedId int
	err := businessDB.DB.QueryRow(insertStatement, 
		individual.FirstName, individual.LastName, 
		individual.Mobile.DialingCode, individual.Mobile.Number, 
		dob,
		individual.Address.Line1, individual.Address.Line2, individual.Address.PostalCode, 
		individual.Address.City,
	).Scan(&insertedId)
	
	if err != nil {
		return nil, err
	}
	
	return &insertedId, nil
}
