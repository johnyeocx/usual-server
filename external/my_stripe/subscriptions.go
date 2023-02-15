package my_stripe

import (
	"time"

	"github.com/johnyeocx/usual/server/db/models"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/subscription"
)




func CreateSubscription(
	customerId string, 
	businessId string,
	cardId		string,
	subProduct models.SubscriptionProduct,
) (*stripe.Subscription, error) {
	stripe.Key = stripeSecretKey()

	// for each item, create subitem params
	items := []*stripe.SubscriptionItemsParams{}
	items = append(items, &stripe.SubscriptionItemsParams{
		Price: subProduct.SubPlan.StripePriceID,
	})

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerId),
		Items: items,

		TransferData: &stripe.SubscriptionTransferDataParams{
			Destination: stripe.String(businessId),
		},
		DefaultPaymentMethod: stripe.String(cardId),
		CollectionMethod: stripe.String("charge_automatically"),
	};
	params.AddExpand("latest_invoice.payment_intent")

	
	s, err := subscription.New(params);
	if err != nil {
		return nil, err
	}
	return s, nil
}

func UpdateSubDefaultCardAndConfirm(
	subId string, 
	cardId string,
) (*stripe.Subscription, *stripe.PaymentIntent, error)  {
	stripe.Key = stripeSecretKey()
	
	params := &stripe.SubscriptionParams{
		DefaultPaymentMethod: stripe.String(cardId),
	}
	params.AddExpand("latest_invoice.payment_intent")

	s, err := subscription.Update(subId, params)
	if err != nil {
		return nil,nil,  err
	}

	p, err := paymentintent.Confirm(s.LatestInvoice.PaymentIntent.ID, &stripe.PaymentIntentConfirmParams{
		PaymentMethod: &cardId,
	})
	if err != nil {
		return nil, nil, err
	}

	return s, p, nil
}

func GetSubLastInvoicePaymentIntent(
	pmId string, 
) (*stripe.PaymentIntent, error)  {
	stripe.Key = stripeSecretKey()


	

	pm, err := paymentintent.Get(pmId, nil)

	if err != nil {
		return nil, err
	}

	return pm, nil
}


func ResumeSubscription(
	cusId string,
	busId string,
	priceId string,
	cardId string,
	expires time.Time,
) (*string, error) {
	stripe.Key = stripeSecretKey()

	// for each item, create subitem params
	items := []*stripe.SubscriptionItemsParams{}
	items = append(items, &stripe.SubscriptionItemsParams{
		Price: stripe.String(priceId),
	})

	var billingAnchor int64
	if (expires.After(time.Now())) {
		billingAnchor = expires.Unix()
	} else {
		billingAnchor = time.Now().Unix()
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(cusId),
		Items: items,

		TransferData: &stripe.SubscriptionTransferDataParams{
			Destination: stripe.String(busId),
		},
		DefaultPaymentMethod: stripe.String(cardId),
		ProrationBehavior: stripe.String("none"),
		// TrialEnd: stripe.Int64(billingAnchor),
		BillingCycleAnchor: stripe.Int64(billingAnchor),
	};
	params.AddExpand("latest_invoice.payment_intent")
	
	s, err := subscription.New(params);
	if err != nil {
		return nil, err
	}
	return &s.ID, nil
}

func CancelSubscription(subId string) (error) {
	stripe.Key = stripeSecretKey()

	_, err := subscription.Cancel(subId, nil)

	return err
}

func ChangeSubDefaultCard(subId string, cardId string) (error) {
	stripe.Key = stripeSecretKey()
	params := stripe.SubscriptionParams{
		DefaultPaymentMethod: stripe.String(cardId),
	}

	_, err := subscription.Update(subId, &params)

	return err
}




