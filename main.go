package main

import (
	"kaps/routes"

	_ "kaps/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			KAPS API document
//	@version		1.0
//	@description	This is a sample server celler server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		192.168.15.248:3000
//	@BasePath	/

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {

	router := gin.Default()
	// router.Use(static.Serve("/", static.LocalFile("./cube/dist", false)))
	// router.NoRoute(func(c *gin.Context) { c.File("./cube/dist/index.html") })

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	public := router.Group("/")
	routes.PublicRoutes(public)

	router.Run("192.168.15.248:3000")
}
