package models

import "time"

type SubInfo struct {
	FirstName 			string 			`json:"first_name"`
	LastName 			string 			`json:"last_name"`
	Email				string 			`json:"email"`
	ProductName			string 			`json:"product_name"`
	ProductID			int 			`json:"product_id"`
	Subscription		Subscription 	`json:"subscription"`
}

type UsageInfo struct {
	CusUUID 		string 		`json:"customer_uuid"`
	CusFirstName 		string 		`json:"cus_first_name"`
	CusLastName 		string 		`json:"cus_last_name"`
	Created 		time.Time	`json:"created"`

	PlanID			int			`json:"plan_id"`
	SubUsage		SubUsage	`json:"sub_usage"`
	ProductID 		int 		`json:"product_id"`
	ProductName 	string 		`json:"product_name"`
	UsageCount		int			`json:"usage_count"`	
}