package models

type Business struct {
	BusinessID 		int `json:"business_id"`
	Name 			string `json:"name"`
	Email 			string `json:"email"`
	Country 		string `json:"country"`
	BusinessCategory 		*string `json:"business_category"`
	BusinessUrl 			*string `json:"business_url"`
	IndividualID 			*int `json:"individual_id"`
	StripeID 		*string `json:"stripe_account_id"`
	Description 	*string `json:"description"`
}

type Person struct {
	ID				int 		`json:"id"`
	FirstName		string 		`json:"first_name"`
	LastName		string 		`json:"last_name"`
	Email			string 		`json:"email"`
	Mobile			PhoneNumber `json:"mobile"`
	DOB				Date 		`json:"dob"`
	Address			Address		`json:"address"`
}

type Date struct {
	Day	int `json:"day"`
	Month	int `json:"month"`
	Year	int `json:"year"`
}

