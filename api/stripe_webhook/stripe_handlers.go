package stripe_webhook

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

		payload, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Println(err)
			return
		}

		event := stripe.Event{}
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Println(err)
			return
		}
		fmt.Println(event)

		// handle invoice.paid event
	}
}