package models

import "time"


type CusUsage struct {
	ID				int				`json:"usage_id"`
	CusUUID			string 			`json:"customer_uuid"`
	Created 		time.Time 		`json:"created"`
	SubUsageID 		int 			`json:"sub_usage_id"`
	

	// FOR CUSTOMER
	SubUsageTitle 		*string 	`json:"title"`
	ProductID 		*int			`json:"product_id"`
	ProductName 	*string			`json:"product_name"`
}
