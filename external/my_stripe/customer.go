package my_stripe

import (
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/paymentmethod"
)

func CreateCustomerNoPayment(newC *models.Customer) (*string, error) {
	stripe.Key = stripeSecretKey()

	name := newC.FirstName + newC.LastName
	// 2. CREATE CUSTOMER
	params := &stripe.CustomerParams{
		Name: stripe.String(name),
		Email: &newC.Email,
	}
	
	c, err := customer.New(params)
	if err != nil {
		return nil, err
	}

	return &c.ID, nil
}

func CreateCustomer(newC *models.Customer, card *models.CreditCard) (*string, error) {
	stripe.Key = stripeSecretKey()

	// 1. CREATE PAYMENT METHOD
	pm, err := CreatePaymentMethod(card)
	if err != nil {
		return nil, err
	}

	// 2. CREATE CUSTOMER
	name := newC.FirstName + " " + newC.LastName
	params := &stripe.CustomerParams{
		Name: stripe.String(name),
		Email: &newC.Email,
		PaymentMethod: &pm.ID,
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: &pm.ID,
		},
	}

	c, err := customer.New(params)
	if err != nil {
		return nil, err
	}

	return &c.ID, nil
}

func CreatePaymentMethod(card *models.CreditCard) (*stripe.PaymentMethod, error) {
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

	return pm, nil
}

func AddNewCustomerCard(cusId string, card *models.CreditCard) (*stripe.PaymentMethod, error) {
	stripe.Key = stripeSecretKey()

	// 1. CREATE PAYMENT METHOD
	pm, err := CreatePaymentMethod(card)
	if err != nil {
		return nil, err
	}

	attachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(cusId),
	  }
	_, err = paymentmethod.Attach(
		pm.ID,
		attachParams,
	)
	if err != nil {
		return nil, err
	}

	// 2. CREATE CUSTOMER
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: &pm.ID,
		},
	}

	_, err = customer.Update(cusId, params)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

func UpdateCusName(cusId string, name string) (error) {
	stripe.Key = stripeSecretKey()

	// 2. CREATE CUSTOMER
	params := &stripe.CustomerParams{
		Name: stripe.String(name),
	}

	_, err := customer.Update(cusId, params)
	return err
}


func UpdateCusEmail(cusId string, email string) (error) {
	stripe.Key = stripeSecretKey()

	// 2. CREATE CUSTOMER
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}

	_, err := customer.Update(cusId, params)
	return err
}

func UpdateCusAddress(cusId string, address models.Address) (error) {
	stripe.Key = stripeSecretKey()

	// 2. CREATE CUSTOMER
	params := &stripe.CustomerParams{
		Address: &stripe.AddressParams{
			Line1: stripe.String(address.Line1),
			Line2: stripe.String(address.Line2),
			PostalCode: stripe.String(address.PostalCode),
			City: stripe.String(address.City),
			Country: address.Country,
		},
	}

	_, err := customer.Update(cusId, params)
	return err
}

