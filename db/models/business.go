package models

type Business struct {
	ID 		int `json:"business_id"`
	Name 			string `json:"name"`
	Email 			string `json:"email"`
	Country 		string `json:"country"`
	BusinessCategory 		*string `json:"business_category"`
	BusinessUrl 			*string `json:"business_url"`
	IndividualID 			*int `json:"individual_id"`
	StripeID 		*string `json:"stripe_account_id"`
	Description 	*string `json:"description"`
	SubCount 		*int	`json:"sub_count"`
}
