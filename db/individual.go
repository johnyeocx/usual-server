package db

import (
	"time"

	"github.com/johnyeocx/usual/server/db/models"
)

// GET
func (b *BusinessDB) GetIndividualID (businessId int) (*int, error) {
	var indivId int
	err := b.DB.QueryRow(`SELECT individual_id FROM business WHERE business_id=$1`,
		businessId,
	).Scan(&indivId)

	if err != nil {
		return nil, err
	}

	return &indivId, nil
}

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

func (businessDB *BusinessDB) GetIndividualByID(individualId int) (*models.Person, error) {
	selectStatement := `SELECT 
		individual_id, first_name, last_name, 
		dialing_code, mobile_number, address_line1, address_line2,
		postal_code, city, dob from individual WHERE individual_id=$1
	`

	individual := models.Person{}

	var dob time.Time
	err := businessDB.DB.QueryRow(selectStatement, individualId).Scan(
		&individual.ID,
		&individual.FirstName,
		&individual.LastName,
		&individual.Mobile.DialingCode,
		&individual.Mobile.Number,
		&individual.Address.Line1,
		&individual.Address.Line2,
		&individual.Address.PostalCode,
		&individual.Address.City,
		&dob,
	)

	if err != nil {
		return nil, err
	}

	individual.DOB.Day = dob.Day()
	individual.DOB.Month = int(dob.Month())
	individual.DOB.Year = dob.Year()

	return &individual, nil
}

func (b *BusinessDB) SetIndividualName(
	individualId int,
	firstName string,
	lastName string,
) (error) {
	_, err := b.DB.Exec(`UPDATE individual SET first_name=$1, last_name=$2 WHERE individual_id=$3`, 
		firstName, lastName, individualId,
	)

	return err
}

func (b *BusinessDB) SetIndividualDOB(
	individualId int,
	day int, 
	month int,
	year int,
) (error) {
	

	dob := time.Date(
		year, time.Month(month), day, 
		0, 0, 0, 0, time.UTC,
	)

	_, err := b.DB.Exec(`UPDATE individual SET dob=$1 WHERE individual_id=$2`, 
		dob, individualId,
	)

	return err
}

func (b *BusinessDB) SetIndividualAddress(
	individualId int,
	line1 string,
	line2 string,
	postalCode string,
	city string,
) (error) {

	_, err := b.DB.Exec(`UPDATE individual SET address_line1=$1, address_line2=$2, postal_code=$3, city=$4 
		WHERE individual_id=$5`, 
		line1, line2, postalCode, city, individualId,
	)

	return err
}

func (b *BusinessDB) SetIndividualMobile(
	individualId int,
	dialingCode string,
	number string,
) (error) {

	_, err := b.DB.Exec(`UPDATE individual SET dialing_code=$1, mobile_number=$2
		WHERE individual_id=$3`, 
		dialingCode, number, individualId,
	)

	return err
}