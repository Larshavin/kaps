package kaas

import (
	"encoding/base64"
	"fmt"
	"kaps/openstack"
	"kaps/types"
	"kaps/utils"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// postOpenstackServerCreateHandler godoc
//
//	@Summary		Create openstack Server
//	@Description	Create openstack Server
//	@Tags			server_create
//	@Accept			json
//	@Produce		json
//	@Param			request	body types.KaasCreateRequest	true	"cluster info"
//	@Param			id	path	string	true	"project id"
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/api/kaas/{id} [post]
func PostKaaSHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		projectID := c.Param("id")
		if projectID == "" {
			panic("project id is empty")
		}

		var kaasRequest types.KaasCreateRequest
		err := c.BindJSON(&kaasRequest)
		if err != nil {
			panic(err)
		}
		fmt.Println(kaasRequest)

		// 0. Connect Websocket to Client

		// 1. Save Nodes information in DB (MongoDB) & check the main control-plane node

		// 1.1 Read the contents of the shell script file
		// Get the path to the main.go file
		mainPath, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		scriptFile := filepath.Join(filepath.Dir(mainPath), "kaps/k8s_provisioning/k8s_control.sh")
		fmt.Println(scriptFile)

		scriptContents, err := os.ReadFile(scriptFile)
		if err != nil {
			panic(err)
		}

		// 1.2: change shell script file contents
		changeVar := map[string]string{
			"control_plane_ip": kaasRequest.Nodes[0].FixedIP,
			"k8s_network_cidr": "172.16.0.0/16",
		}
		updatedScript, err := utils.FixShellScriptVariables(string(scriptContents), changeVar)
		if err != nil {
			panic(err)
		}

		//  updatedScript must be Base64 encoded. Restricted to 65535 bytes.
		encodedScript := base64.StdEncoding.EncodeToString([]byte(updatedScript))

		// 2. Create keypair and save private key or use existing keypair by keypair name
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

		// response error가 409일 때 이미 존재하는 keypair이므로, 그때는 keypair를 생성하지 않고 그대로 사용한다.
		keypair, err := openstack.CreateKeypair(tokenID, kaasRequest.KeyName)
		if err != nil {
			panic(err)
		}

		// 3. Create the main k8s control-Plane node on openstack
		randomString, err := utils.GenerateRandomString(8)
		if err != nil {
			panic(err)
		}
		var mainControlPlaneNode types.ServerCreateRequest
		mainControlPlaneNode.Server.Name = kaasRequest.Name + "-cp-" + randomString
		mainControlPlaneNode.Server.Flavor = "3"
		mainControlPlaneNode.Server.Networks = []types.ServerCreateNetwork{
			{
				UUID:    kaasRequest.NetworkId,
				FixedIP: &kaasRequest.Nodes[0].FixedIP,
			},
			{
				UUID:    "abe6e177-dd45-4923-9cbd-ae2a09ce4fb9", // mgmt : "abe6e177-dd45-4923-9cbd-ae2a09ce4fb9", provider :"d64dbfc6-46de-4f2c-8b65-12c9d29a8b7e",
				FixedIP: nil,
			},
		}
		mainControlPlaneNode.Server.SecurityGroups = []types.ServerCreateSecurityGroup{
			{
				Name: "default",
			},
			{
				Name: "k8s-common",
			},
			{
				Name: "k8s-control-plane",
			},
		}
		mainControlPlaneNode.Server.KeyName = kaasRequest.KeyName
		mainControlPlaneNode.Server.Image = "e62a8d62-02f9-4bbd-ae13-ca298faec579" // image : k8s-centos8-openstack
		mainControlPlaneNode.Server.Script = encodedScript

		responseOfServerCreation, err := openstack.CreateServer(tokenID, projectID, mainControlPlaneNode)
		if err != nil {
			panic(err)
		}

		server, ok := responseOfServerCreation["server"].(map[string]interface{})
		if !ok {
			fmt.Println("Invalid server data")
			panic(ok)
		}

		id, ok := server["id"].(string)
		if !ok {
			fmt.Println("Invalid id")
			panic(ok)
		}

		var fixedMgmtIP, status string
		timeout := time.After(60 * time.Second) // Set the timeout to 60 seconds

	OuterLoopForVMStatus:
		for fixedMgmtIP == "" && status != "ACTIVE" {
			select {
			case <-timeout:
				fmt.Println("Timeout reached")
				break OuterLoopForVMStatus // Exit the outer loop if timeout is reached
			default:
				createVMdetail, err := openstack.GetServerDetail(tokenID, id)
				if err != nil {
					panic(err)
				}
				if createVMdetail.Server.Status == "ACTIVE" {
					for k, v := range createVMdetail.Server.Addresses {
						if k == "mgmt" {
							fixedMgmtIP = v[0].Addr
							break OuterLoopForVMStatus // Exit the outer loop if fixedMgmtIP is obtained
						}
					}
				}
				time.Sleep(1 * time.Second)
			}
		}

		// 4. Wait for the main k8s control-Plane node to be ready
		// 	  Retry to get cluster info from the main k8s control-Plane node by SSH
		//    And save the k8s cluster's hash & token data in DB (MongoDB)

		keyFile := filepath.Join(filepath.Dir(mainPath), "ssh/k8s.pem")

		sshClient := utils.SSH{
			IP:   fixedMgmtIP, // use Mgmt IP, but it is not known yet.
			User: "centos",
			Port: 22,
			Cert: keyFile, // Password or Key file path
		}

		var kubeadmJoinOutput string
		timeout = time.After(120 * time.Second) // Set the timeout to 120 seconds

	OuterLoopForSSH:
		for kubeadmJoinOutput == "" {
			select {
			case <-timeout:
				fmt.Println("Timeout reached")
				break OuterLoopForSSH // Exit the outer loop if timeout is reached
			default:
				kubeadmJoinOutput, err = utils.GetKubeadmJoinOutput(sshClient, utils.CertPublicKeyFile)
				if err != nil {
					fmt.Println(err)
					break
				}
				fmt.Println(kubeadmJoinOutput)
				time.Sleep(1 * time.Second)
			}
		}

		// 5. Make the Data-plane nodes on openstack

		// fmt.Println(users)

		c.JSON(http.StatusOK, gin.H{
			"content":  kaasRequest,
			"keypair":  keypair,
			"response": responseOfServerCreation,
			"cmd":      kubeadmJoinOutput,
		})
	}
}
