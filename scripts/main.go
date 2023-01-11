package main

import (
	"log"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
)

func DeleteConnectedAccounts(
	) () {
		stripe.Key = "sk_test_51MBuhlBVuVjaXA7P2WAxmTbhF1BRrFUNg5ZUMygVyebItDQVqNHiUz5kJYfBG6dSdcGTidCNQBnrC12cRScvosns004o5lJPQO"
	
		params := &stripe.AccountListParams{}
		params.Filters.AddFilter("limit", "", "3")
		i := account.List(params)
		for i.Next() {
			a := i.Account()
			_, err := account.Del(a.ID, nil)
			if err != nil {
				log.Printf("Failed to delete account\n")
			}
	
		}
	}
	
func main () {
	DeleteConnectedAccounts()
}