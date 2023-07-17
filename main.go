package main

import (
	"context"
	"kaps/routes"
	"kaps/types"

	_ "kaps/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const uri = "mongodb://192.168.15.47:27017"

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

	// Initialize the MongoDB variable
	MongoDB := &types.MongoDB{}
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	// Assign the client to the MongoDB.Client field
	MongoDB.Client = client
	defer func() {
		if err = MongoDB.Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	router := gin.Default()
	// router.Use(static.Serve("/", static.LocalFile("./cube/dist", false)))
	// router.NoRoute(func(c *gin.Context) { c.File("./cube/dist/index.html") })
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	public := router.Group("/")
	routes.PublicRoutes(public, MongoDB)
	router.Run("192.168.15.248:3000")
}
