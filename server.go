package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/johnyeocx/usual/server/routes"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/joho/godotenv"
)

func main() {
	
	// return

	// 1. Load env file
	err := godotenv.Load(".env")
	if err != nil {
		return 
	}

	// my_stripe.DeleteAllStripeProducts()
	// return

	// 2. Connect to DB
	psqlDB := db.Connect()

	// 3. Run cron jobs
	// go scheduled.RunCronJobs(psqlDB)

	// 4. Connect to S3
	sess := cloud.ConnectAWS()
	// media.GenerateSubscribeQRCode(sess, 1)
	// return

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "DELETE", "POST",},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))

	router.Use(middleware.AuthMiddleware())
	
	// router.Use(utils.AuthMiddleware())
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, "Welcome to the usual api")
	})
	routes.CreateRoutes(router, psqlDB, sess)
	
	router.Run(":8080")
}