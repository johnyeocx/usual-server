package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/johnyeocx/usual/server/routes"
	"github.com/johnyeocx/usual/server/utils/fcm"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/joho/godotenv"
)

func main() {
	
	
	

	// 1. Load env file
	err := godotenv.Load(".env")
	if err != nil {
		return 
	}


	
	// 2. Connect to services
	psqlDB := db.Connect()
	sess := cloud.ConnectAWS()
	fbApp, err := fcm.CreateFirebaseApp()
	if err != nil {
		panic(err);
	}

	// 3. Run cron jobs
	// go scheduled.RunCronJobs(psqlDB)
	
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		// AllowOrigins:     []string{"http://172.28.38.241:3000", "http://172.28.38.241:3001"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "DELETE", "POST",},
		AllowHeaders:     []string{"Content-Type", "Authorization", "*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))

	router.Use(middleware.AuthMiddleware())
	
	// router.Use(utils.AuthMiddleware())
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, "Welcome to the usual api")
	})
	routes.CreateRoutes(router, psqlDB, sess, fbApp)
	
	router.Run(":8080")
}