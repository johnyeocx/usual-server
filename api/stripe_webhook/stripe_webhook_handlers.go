package stripe_webhook

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
)

func Routes(stripeWRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	stripeWRouter.POST("", stripeWebhookHandler(sqlDB))
}

func stripeWebhookHandler(sqlDB *sql.DB) gin.HandlerFunc {
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
			err := InsertInvoice(sqlDB, event.Data.Object, "succeeded")
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
			err := InsertInvoice(sqlDB, event.Data.Object, "requires_action")
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
			// REST
			err := InsertInvoice(sqlDB, event.Data.Object, "payment_failed")
			if err != nil {
				log.Println("Failed to insert invoice payment failed:", err)
				c.JSON(http.StatusBadGateway, err)
				return
			} else {
				c.JSON(200, nil)
				return
			}
			// SEND NOTIFICATION TO CUSTOMER_ID
		}
	}
}