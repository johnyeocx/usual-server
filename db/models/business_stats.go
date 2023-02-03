package models

import "time"

type SubInfo struct {
	CusName 			string 			`json:"customer_name"`
	ProductName			string 			`json:"product_name"`
	ProductID			int 			`json:"product_id"`
	Subscription		Subscription 	`json:"subscription"`
}

type UsageInfo struct {
	CusUUID 		string 		`json:"customer_uuid"`
	CusName 		string 		`json:"customer_name"`
	Created 		time.Time	`json:"created"`

	PlanID			int			`json:"plan_id"`
	SubUsage		SubUsage	`json:"sub_usage"`
	ProductID 		int 		`json:"product_id"`
	ProductName 	string 		`json:"product_name"`
	UsageCount		int			`json:"usage_count"`	
}