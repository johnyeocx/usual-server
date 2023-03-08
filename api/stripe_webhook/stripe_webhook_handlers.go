package stripe_webhook

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	sw_payout "github.com/johnyeocx/usual/server/api/stripe_webhook/business_payout"
	constants "github.com/johnyeocx/usual/server/constants/enums"
	"github.com/stripe/stripe-go/v74"
)

func Routes(
	stripeWRouter *gin.RouterGroup, 
	sqlDB *sql.DB, 
	s3Sess *session.Session, 
	firebaseApp *firebase.App,
) {
	stripeWRouter.POST("", stripeWebhookHandler(sqlDB, firebaseApp))
}

func stripeWebhookHandler(sqlDB *sql.DB, firebaseApp *firebase.App) gin.HandlerFunc {
	return func (c *gin.Context) {
		
		const MaxBodyBytes = int64(65536)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		event := stripe.Event{}
		if err := json.Unmarshal(payload, &event); err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}	
		
		switch event.Type {

		case "invoice.paid": 
			_, err := InsertInvoice(sqlDB, firebaseApp, event.Data.Object, constants.PMIPaymentSucceeded)
			if err != nil {
				log.Println("Failed to insert invoice paid:", err)
				c.JSON(http.StatusBadGateway, err)
				return
			} else {
				c.JSON(200, nil)
				return
			}
		
		case "invoice.payment_action_required":
			_, err := InsertInvoice(sqlDB, firebaseApp, event.Data.Object, constants.PMIPaymentRequiresAction)
			if err != nil {
				log.Println("Failed to insert invoice action required:", err)
				c.JSON(http.StatusBadGateway, err)
				return
			} else {
				c.JSON(200, nil)
				return
			}
		case "invoice.payment_failed": 
			_, err := InsertInvoice(sqlDB, firebaseApp, event.Data.Object, constants.PMIPaymentFailed)
				

			if err != nil {
				log.Println("Failed to insert invoice payment failed:", err)
				c.JSON(http.StatusBadGateway, err)
				return
			}
		
		case "account.updated":
			var updatedAccount stripe.Account
			err := json.Unmarshal(event.Data.Raw, &updatedAccount)
			if err != nil {
				log.Println("Error parsing JSON:", err)
				c.JSON(http.StatusBadRequest, err)
				return
			}

			currentlyDue := updatedAccount.Requirements.CurrentlyDue
			// currentlyDue = append(currentlyDue, "abc.verification.document")
			// updatedAccount.ID = "acct_1MhxGVPiTrrt6DZt"

			if (VerificationDocRequired(currentlyDue)) {
				reqErr := SetIndVerificationDocRequired(sqlDB, updatedAccount, true)
				if reqErr != nil {
					log.Println("Failed to set ind verification doc required:", reqErr.Err)
					c.JSON(reqErr.StatusCode, reqErr.Err)
					return
				}
			} 
			c.JSON(200, nil)
		
		case "payout.paid":
			var payout stripe.Payout
			err := json.Unmarshal(event.Data.Raw, &payout)
			if err != nil {
				log.Println("Error parsing JSON:", err)
				c.JSON(http.StatusBadRequest, err)
				return
			}
			// payout.Amount

			reqErr := sw_payout.InsertPayout(sqlDB, payout)
			if reqErr != nil {
				reqErr.Log()
				c.JSON(reqErr.StatusCode, reqErr.Err)
				return
			}
			c.String(http.StatusOK, "Payout inserted successfully")
		}
	}
}

