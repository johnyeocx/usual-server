package fcm

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	my_enums "github.com/johnyeocx/usual/server/constants/enums"
	"google.golang.org/api/option"
)

func CreateFirebaseApp() (*firebase.App, error){
	opt := option.WithCredentialsFile("./credentials/firebase_admin_credentials.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return app, nil
}

func SendPaymentFailedNotification(
	app *firebase.App, 
	fcmToken string,
	subId int,
	productName string,
	businessName string,
) (error){

	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		return err
	}

	msgBody := fmt.Sprintf("You have a pending payment for your subscription to %s by %s", productName, businessName)
	_, err = fcmClient.Send(context.Background(), &messaging.Message{
		Notification: &messaging.Notification{
		  Title: "Payment Pending",
		  Body: msgBody,
		},

		Token: fcmToken, 
		Data: map[string]string{
			"type": string(my_enums.PNPaymentFailed),
			"sub_id": fmt.Sprint(subId),
			
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
		},
	})
	
	return err
}

func SendPaymentSucceededNotification(
	app *firebase.App, 
	fcmToken string,
	subId int,
	cost int,
	productName string,
	businessName string,
) (error){

	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		return err
	}

	costAsFloat := float64(cost/100)

	msgBody := fmt.Sprintf("You have successfully paid £%.2f for your subscription to %s by %s", costAsFloat, productName, businessName)
	_, err = fcmClient.Send(context.Background(), &messaging.Message{
		Notification: &messaging.Notification{
		  Title: "Payment Succeeded",
		  Body: msgBody,
		},

		Token: fcmToken, 
		Data: map[string]string{
			"type": string(my_enums.PNPaymentRequiresAction),
			"sub_id": fmt.Sprint(subId),
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
		},
	})
	
	return err
}

func SendPaymentVoidedNotification(
	app *firebase.App, 
	fcmToken string,
	subId int,
	productName string,
	businessName string,
) (error){

	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		return err
	}

	msgBody := fmt.Sprintf("Your subscription to %s by %s has been cancelled due to an expired payment", productName, businessName)
	_, err = fcmClient.Send(context.Background(), &messaging.Message{
		Notification: &messaging.Notification{
		  Title: "Payment Failed",
		  Body: msgBody,
		},

		Token: fcmToken, 
		Data: map[string]string{
			"type": string(my_enums.PNSubCancelled),
			"sub_id": fmt.Sprint(subId),
			
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
		},
	})
	
	return err
}
