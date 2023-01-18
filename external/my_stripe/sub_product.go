package my_stripe

import (
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
)

func CreateNewSubProduct(
	productName string, 
	plan models.SubscriptionPlan,
) (*string, *string, error) {
	stripe.Key = stripeSecretKey()

	
	params := &stripe.PriceParams{
		ProductData: &stripe.PriceProductDataParams{
			Name: stripe.String(productName),
		},
		Currency: stripe.String(plan.Currency),
		UnitAmount: stripe.Int64(int64(plan.UnitAmount)),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(plan.RecurringDuration.Interval.String),
			IntervalCount: stripe.Int64(int64(plan.RecurringDuration.IntervalCount.Int16)),
		},
	}

	p, err := price.New(params)
	if err != nil {
		return nil, nil, err
	}

	return &p.Product.ID, &p.ID, nil
}


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


func UpdateSubProductPrice(
	priceId string,
	productId string, 
	recurringDuration models.TimeFrame,
	unitAmount int,
) (*string, error) {
	stripe.Key = stripeSecretKey()
	
	// archive new price
	params := &stripe.PriceParams{
		Active: stripe.Bool(false),
	}

	params = &stripe.PriceParams{
		Product: &productId,
		Currency: stripe.String("GBP"),
		UnitAmount: stripe.Int64(int64(unitAmount)),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(recurringDuration.Interval.String),
			IntervalCount: stripe.Int64(int64(recurringDuration.IntervalCount.Int16)),
		},
	}

	p, err := price.New(params)
	if err != nil {
		return nil, err
	}
	return &p.ID, nil
}


func UpdateProductName(
	productId string, 
	name string,
) (error) {
	stripe.Key = stripeSecretKey()

	params := &stripe.ProductParams{
		Name: stripe.String(name),
	}

	_, err := product.Update(productId, params)
	return err
}

func DeleteAllStripeProducts () {

	stripe.Key = stripeSecretKey()

	params := &stripe.ProductListParams{}
	i := product.List(params)
	for i.Next() {
		p := i.Product()
		_, _ = product.Del(p.ID, nil)
	}
}

