package my_stripe

import (
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/paymentmethod"
)

func CreateCustomer(newC *models.Customer, card *models.CreditCard) (*string, error) {
	stripe.Key = stripeSecretKey()

	// 1. CREATE PAYMENT METHOD
	paymentMethod, err := CreatePaymentMethod(card)
	if err != nil {
		return nil, err
	}

	// 2. CREATE CUSTOMER
	params := &stripe.CustomerParams{
		Name: &newC.Name,
		Email: &newC.Email,
		PaymentMethod: paymentMethod,
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: paymentMethod,
		},
		
	}
	c, err := customer.New(params)
	if err != nil {
		return nil, err
	}

	return &c.ID, nil
}

func CreatePaymentMethod(card *models.CreditCard) (*string, error) {
	stripe.Key = stripeSecretKey()

	params := &stripe.PaymentMethodParams{
		Card: &stripe.PaymentMethodCardParams{
		  Number: stripe.String(card.Number),
		  ExpMonth: stripe.Int64(card.ExpMonth),
		  ExpYear: stripe.Int64(card.ExpYear),
		  CVC: stripe.String(card.CVC),
		},
		Type: stripe.String("card"),
		
	}

	pm, err := paymentmethod.New(params)

	if err != nil {
		return nil, err
	}

	return &pm.ID, nil
}