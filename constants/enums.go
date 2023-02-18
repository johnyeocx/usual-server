package constants

import "github.com/stripe/stripe-go/v74"


type PushNotificationType string

const (
	PNPaymentFailed					PushNotificationType = "payment_failed"
	PNPaymentRequiresAction        	PushNotificationType = "payment_requires_action"
	PNPaymentSucceeded             	PushNotificationType = "payment_succeeded"


	PNSubCancelled	             	PushNotificationType = "subscription_cancelled"
)


type MyPaymentIntentStatus string

const (
	PMIPaymentFailed        	MyPaymentIntentStatus = "payment_failed"
	PMIPaymentRequiresAction    MyPaymentIntentStatus = "requires_action"
	PMIPaymentSucceeded        	MyPaymentIntentStatus = "succeeded"
	PMIPaymentCancelled        	MyPaymentIntentStatus = "cancelled"
)

func StripePMStatusToMYPMStatus(status stripe.PaymentIntentStatus) (MyPaymentIntentStatus) {
	if status == stripe.PaymentIntentStatusSucceeded {
		return PMIPaymentSucceeded
	} else if status == stripe.PaymentIntentStatusRequiresAction {
		return PMIPaymentRequiresAction
	} else {
		return PMIPaymentFailed
	}
}