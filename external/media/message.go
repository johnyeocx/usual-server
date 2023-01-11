package media

import (
	"fmt"

	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

func MessageClient(msg string, toMobile string) (error) {

	client := twilio.NewRestClient()
	
    params := &openapi.CreateMessageParams{}
    params.SetTo(toMobile)
    params.SetFrom("usual")

    params.SetBody(msg)

    _, err := client.Api.CreateMessage(params)
    if err != nil {
        return fmt.Errorf("failed to message otp to client: \n%v", err)
    }

    return nil
}