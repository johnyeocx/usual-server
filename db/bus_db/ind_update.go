package busdb

import "time"




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

func (b *BusinessDB) UpdateIndividualVerificationDocumentRequired(
	individualId int,
	required bool,
) (error) {

	_, err := b.DB.Exec(`UPDATE individual SET verification_document_required=$1
		WHERE individual_id=$2`, 
		required, individualId,
	)

	return err
}