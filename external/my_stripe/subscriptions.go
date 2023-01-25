package my_stripe

import (
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/subscription"
)



func CreateSubscription(
	customerId string, 
	businessId string,
	cardId		string,
	subProduct models.SubscriptionProduct,
) (*string, error) {
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
	return &s.ID, nil
}

func CancelSubscription(subId string) (error) {
	stripe.Key = stripeSecretKey()

	_, err := subscription.Cancel(
		subId,
		nil,
	)

	return err
}



