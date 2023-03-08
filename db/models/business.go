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
	EmailVerified 	*bool 	`json:"email_verified"`
	ExternalAccountID JsonNullInt16 	`json:"external_account_id"`
	ExternalAccountType JsonNullString 	`json:"external_account_type"`
}

type BankAccount struct {
	ID					int 	`json:"bank_account_id"`
	StripeID 			string 	`json:"stripe_id"`
	AccountHolderName 	string 	`json:"account_holder_name"`
	// AccountHolderType 	string 	`json:"account_holder_type"`
	BankName			string 	`json:"bank_name"`
	Last4				string 	`json:"last4"`
	RoutingNumber 		string 	`json:"routing_number"`
}

type BankInfo struct {
	AccountHolder 	string 	`json:"account_holder"`
	AccountNumber 	string 	`json:"account_number"`
	RoutingNumber		string 	`json:"routing_number"`
}
