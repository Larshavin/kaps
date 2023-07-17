package routes

import (
	"fmt"
	"kaps/kaas"
	"kaps/openstack"
	"kaps/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PublicRoutes(g *gin.RouterGroup, mongoDB *types.MongoDB) {
	g.GET("/api/", IndexGetHandler())
	g.GET("/api/user", getOpenstackUserHandler())
	g.GET("/api/server", getOpensStackServerListHandler())
	g.GET("/api/server/:id", getOpensStackServerDetailHandler())
	g.GET("/api/image", getOpenStackImageListHandler())
	g.GET("/api/network", getOpenStackNetworkListHandler())
	g.POST("/api/server/:id", postOpenstackServerCreateHandler())
	g.POST("/api/kaas/:id", kaas.PostKaaSHandler(mongoDB))
}

func IndexGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"content": "This is an index page...",
		})
	}
}

func getOpenstackUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		identity := types.Identity{
			Username:   "admin",
			Password:   "1234qwer",
			DomainName: "default",
			ProjectId:  "a8e14f37c44d460788ce9fa9d825b5c9",
		}

		tokenID, err := openstack.RequestKeystoneToken(identity)
		if err != nil {
			panic(err)
		}

		users, err := openstack.GetUsersList(tokenID)
		if err != nil {
			panic(err)
		}

		fmt.Println(users)

		c.JSON(http.StatusOK, gin.H{
			"content": users,
		})
	}
}

func getOpensStackServerListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		identity := types.Identity{
			Username:   "admin",
			Password:   "1234qwer",
			DomainName: "default",
			ProjectId:  "a8e14f37c44d460788ce9fa9d825b5c9",
		}

		tokenID, err := openstack.RequestKeystoneToken(identity)
		if err != nil {
			panic(err)
		}

		servers, err := openstack.GetServersList(tokenID)
		if err != nil {
			panic(err)
		}

		fmt.Println(servers)

		c.JSON(http.StatusOK, gin.H{
			"content": servers,
		})
	}
}

func getOpensStackServerDetailHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		identity := types.Identity{
			Username:   "admin",
			Password:   "1234qwer",
			DomainName: "default",
			ProjectId:  "a8e14f37c44d460788ce9fa9d825b5c9",
		}

		tokenID, err := openstack.RequestKeystoneToken(identity)
		if err != nil {
			panic(err)
		}

		serverID := c.Param("id")
		serverDetail, err := openstack.GetServerDetail(tokenID, serverID)
		if err != nil {
			panic(err)
		}

		fmt.Println(serverDetail)

		c.JSON(http.StatusOK, gin.H{
			"content": serverDetail,
		})
	}
}

func getOpenStackImageListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := types.Identity{
			Username:   "admin",
			Password:   "1234qwer",
			DomainName: "default",
			ProjectId:  "a8e14f37c44d460788ce9fa9d825b5c9",
		}

		tokenID, err := openstack.RequestKeystoneToken(identity)
		if err != nil {
			panic(err)
		}

		imageList, err := openstack.GetImageList(tokenID)
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"content": imageList,
		})
	}
}

// getOpenStackNetworkListHandler godoc
//
//	@Summary		Network List
//	@Description	get Network List
//	@Tags			networks
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/api/network [get]
func getOpenStackNetworkListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		identity := types.Identity{
			Username:   "admin",
			Password:   "1234qwer",
			DomainName: "default",
			ProjectId:  "a8e14f37c44d460788ce9fa9d825b5c9",
		}

		tokenID, err := openstack.RequestKeystoneToken(identity)
		if err != nil {
			panic(err)
		}

		networkList, err := openstack.GetNetworkList(tokenID)
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"content": networkList,
		})
	}
}

// postOpenstackServerCreateHandler godoc
//
//	@Summary		Create openstack Server
//	@Description	Create openstack Server
//	@Tags			server_create
//	@Accept			json
//	@Produce		json
//	@Param			request	body	types.ServerCreateRequest	true	"server info"
//
// @Param			id	path	string	true	"project id"
//
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/api/server/{id} [post]
func postOpenstackServerCreateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		if projectID == "" {
			panic("project id is empty")
		}

		var serverCreateRequest types.ServerCreateRequest
		err := c.BindJSON(&serverCreateRequest)
		if err != nil {
			panic(err)
		}

		identity := types.Identity{
			Username:   "admin",
			Password:   "1234qwer",
			DomainName: "default",
			ProjectId:  projectID,
		}

		tokenID, err := openstack.RequestKeystoneToken(identity)
		if err != nil {
			panic(err)
		}

		responseData, err := openstack.CreateServer(tokenID, projectID, serverCreateRequest)
		if err != nil {
			panic(err)
		}

		fmt.Println(responseData)

		c.JSON(http.StatusOK, gin.H{
			"content": responseData,
		})
	}
}
