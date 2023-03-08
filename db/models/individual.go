package models


type Person struct {
	ID				int 		`json:"id"`
	FirstName		string 		`json:"first_name"`
	LastName		string 		`json:"last_name"`
	Email			string 		`json:"email"`
	Mobile			PhoneNumber `json:"mobile"`
	DOB				Date 		`json:"dob"`
	Address			Address		`json:"address"`
	VerificationDocumentRequired bool `json:"verification_document_required"`
}

type Date struct {
	Day	int `json:"day"`
	Month	int `json:"month"`
	Year	int `json:"year"`
}


type Address struct {
	Line1 			string 		`json:"line1"`
	Line2 			string 		`json:"line2"`
	PostalCode 		string 		`json:"postal_code"`
	City 			string 		`json:"city"`
	Country 		*string 	`json:"country"`
}