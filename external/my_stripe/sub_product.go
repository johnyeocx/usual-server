package my_stripe

import (
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/subscription"
)

func DisableProduct(productId string, priceId string) (error) {
	stripe.Key = stripeSecretKey()

	params := &stripe.PriceParams{
		Active: stripe.Bool(false),
	}
	_, err := price.Update(
		priceId,
		params,
	)
	if err != nil {
		return err
	}

	prodParams := &stripe.ProductParams{
		Active: stripe.Bool(false),
	};
	_, err = product.Update(productId, prodParams);
	return err
}

func CancelSubscription(subId string) (error) {
	stripe.Key = stripeSecretKey()

	_, err := subscription.Cancel(
		subId,
		nil,
	)

	return err
}