package restapi

import (
	"fmt"

	docs "github.com/dbschenker/heap-dump-management/heap-dump-service/docs"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/config"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/logging"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/rest-api/auth"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/rest-api/requests"
	apiV1 "github.com/dbschenker/heap-dump-management/heap-dump-service/internal/rest-api/requests/v1"
	"github.com/gin-gonic/gin"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const BASE_PATH = "/api/v1"
const UPLOAD_ENDPOINT = "/upload"

func Serve(cfg *config.AppConfig) {

	docs.SwaggerInfo.BasePath = BASE_PATH
	router := gin.New()
	router.SetTrustedProxies([]string{"10.0.0.0/8"})

	router.Use(logging.JSONLogMiddleware())
	router.Use(gin.Recovery())

	router.Use(func(c *gin.Context) {
		c.Set("cfg", cfg)
		c.Next()
	})

	v1 := router.Group(BASE_PATH)
	{
		v1.POST(UPLOAD_ENDPOINT, auth.SaAuth, apiV1.HandleRequestUpload)
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/health", requests.Health)
	router.GET("/liveness", requests.Liveness)

	router.Run(":" + fmt.Sprint(cfg.App.Port))
}
