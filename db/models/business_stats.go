package models

import "time"

type SubInfo struct {
	CusName 			string 			`json:"customer_name"`
	ProductName			string 			`json:"product_name"`
	ProductID			int 			`json:"product_id"`
	Subscription		Subscription 	`json:"subscription"`
}

type CusUsage struct {
	ID			int			`json:"usage_id"`
	CusUUID		string 		`json:"customer_uuid"`
	Created 	time.Time 	`json:"created"`
	SubUsageID 	int 		`json:"sub_usage_id"`
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