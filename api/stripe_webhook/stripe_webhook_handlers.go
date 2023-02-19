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
			log.Println(err)
			return
		}

		event := stripe.Event{}
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Println(err)
			return
		}	

		if event.Type == "invoice.paid" {
			_, err := InsertInvoice(sqlDB, firebaseApp, event.Data.Object, constants.PMIPaymentSucceeded)
			if err != nil {
				log.Println("Failed to insert invoice paid:", err)
				c.JSON(http.StatusBadGateway, err)
				return
			} else {
				c.JSON(200, nil)
				return
			}
		}

		if event.Type == "invoice.payment_action_required" {
			_, err := InsertInvoice(sqlDB, firebaseApp, event.Data.Object, constants.PMIPaymentRequiresAction)
			if err != nil {
				log.Println("Failed to insert invoice action required:", err)
				c.JSON(http.StatusBadGateway, err)
				return
			} else {
				c.JSON(200, nil)
				return
			}
		}

		if event.Type == "invoice.payment_failed" {
			_, err := InsertInvoice(sqlDB, firebaseApp, event.Data.Object, constants.PMIPaymentFailed)
			

			if err != nil {
				log.Println("Failed to insert invoice payment failed:", err)
				c.JSON(http.StatusBadGateway, err)
				return
			}
		}

		if event.Type == "invoice.voided" {
			err := VoidedInvoice(sqlDB, firebaseApp, event.Data.Object)
			if err != nil {
				log.Println("Failed to void invoice paid:", err)
				c.JSON(http.StatusBadGateway, err)
				return
			} else {
				c.JSON(200, nil)
				return
			}
		}
	}
}