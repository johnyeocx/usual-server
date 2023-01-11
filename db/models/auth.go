package models

type BusinessDetails struct {
	Name 	string `json:"name"`
	Email 	string `json:"email"`
	Country	string `json:"country"`
	Password string `json:"password"`
}

type CreditCard struct {
	Number 		string `json:"number"`
	ExpMonth 	int64 `json:"expiry_month"`
	ExpYear 	int64 `json:"expiry_year"`
	CVC 		string `json:"cvc"`
	Currency 	string `json:"currency"`
}

type PhoneNumber struct {
	DialingCode		string	`json:"dialing_code"`
	Number			string	`json:"number"`
}