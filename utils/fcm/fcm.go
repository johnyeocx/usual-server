package fcm

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/johnyeocx/usual/server/constants"
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

	msgBody := fmt.Sprintf("You have a failed payment for your subscription to %s by %s", productName, businessName)
	_, err = fcmClient.Send(context.Background(), &messaging.Message{
		Notification: &messaging.Notification{
		  Title: "Payment Failed",
		  Body: msgBody,
		},

		Token: fcmToken, 
		Data: map[string]string{
			"type": string(constants.PNPaymentFailed),
			"sub_id": fmt.Sprint(subId),
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

	msgBody := fmt.Sprintf("You have successfully paid £%f for your subscription to %s by %s", costAsFloat, productName, businessName)
	_, err = fcmClient.Send(context.Background(), &messaging.Message{
		Notification: &messaging.Notification{
		  Title: "Payment Succeeded",
		  Body: msgBody,
		},

		Token: fcmToken, 
		Data: map[string]string{
			"type": string(constants.PNPaymentRequiresAction),
			"sub_id": fmt.Sprint(subId),
		},
	})
	
	return err
}

