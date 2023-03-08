package busdb

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

func (businessDB *BusinessDB) GetIndividualByID(individualId int) (*models.Person, error) {
	selectStatement := `SELECT 
		individual_id, first_name, last_name, 
		dialing_code, mobile_number, address_line1, address_line2,
		postal_code, city, dob, verification_document_required from individual WHERE individual_id=$1
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
		&individual.VerificationDocumentRequired,
	)

	if err != nil {
		return nil, err
	}

	individual.DOB.Day = dob.Day()
	individual.DOB.Month = int(dob.Month())
	individual.DOB.Year = dob.Year()

	return &individual, nil
}

func (businessDB *BusinessDB) GetIndividualFromStripeID(busStripeId string) (*models.Person, error) {
	selectStatement := `SELECT 
		i.individual_id 
		FROM business as b JOIN individual as i on b.individual_id=i.individual_id
		WHERE b.stripe_id=$1
	`

	individual := models.Person{}
	err := businessDB.DB.QueryRow(selectStatement, busStripeId).Scan(
		&individual.ID,
	)

	if err != nil {
		return nil, err
	}

	return &individual, nil
}