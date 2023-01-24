package routes

import (
	"database/sql"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/api/auth"
	"github.com/johnyeocx/usual/server/api/business"

	"github.com/johnyeocx/usual/server/api/c/customer"
	"github.com/johnyeocx/usual/server/api/c/subscription"
	"github.com/johnyeocx/usual/server/api/stripe_webhook"
	"github.com/johnyeocx/usual/server/api/sub_product"

	c_auth "github.com/johnyeocx/usual/server/api/c/auth"
	c_business "github.com/johnyeocx/usual/server/api/c/business"
)



func CreateRoutes(router *gin.Engine, db *sql.DB, s3Sess *session.Session) {


	apiRoute := router.Group("/api")
	{
		auth.Routes(apiRoute.Group("/auth"), db, s3Sess)
		business.Routes(apiRoute.Group("/business"), db, s3Sess)
		stripe_webhook.Routes(apiRoute.Group("/stripe_webhook"), db, s3Sess)
		sub_product.Routes(apiRoute.Group("/business/subscription_product"), db, s3Sess)

		c_business.Routes(apiRoute.Group("/c/business"), db, s3Sess)
		customer.Routes(apiRoute.Group("/c/customer"), db, s3Sess)
		c_auth.Routes(apiRoute.Group("/c/auth"), db, s3Sess)
		subscription.Routes(apiRoute.Group("/c/subscription"), db, s3Sess)
	}
}
