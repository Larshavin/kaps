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
	"sync"
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
func PostKaaSHandler(mongoDB *types.MongoDB) gin.HandlerFunc {
	return func(c *gin.Context) {

		projectID := c.Param("id")
		if projectID == "" {
			panic("project id is empty")
		}

		var k8sCluster types.K8SCluster
		var members []types.Node
		var kaasRequest types.KaasCreateRequest
		err := c.BindJSON(&kaasRequest)
		if err != nil {
			panic(err)
		}
		fmt.Println(kaasRequest)

		k8sCluster.Name = kaasRequest.Name
		k8sCluster.Version = kaasRequest.Version
		k8sCluster.NetworkId = kaasRequest.NetworkId
		k8sCluster.KeyName = kaasRequest.KeyName
		k8sCluster.ProjectId = projectID

		// 0. Connect Websocket to Client & setup server create channel

		mainResultChannel := make(chan map[string]interface{})
		resultChannel := make(chan string)

		// 1. Import shell script for Control Plane Node
		// 1.1 Read the contents of the shell script file for main control-plane node
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

		// 1.2: change shell script file contents for main control-plane node
		changeVar := map[string]string{
			"control_plane_ip": kaasRequest.ControlPlaneNodes[0].FixedIP,
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

		// 3. Create the main k8s control-Plane node on openstack (with shell script)
		// 	  And send the result to the resultChannel
		fmt.Println("Create the main k8s control-Plane node on openstack (with shell script)")
		for i := 0; i < len(kaasRequest.ControlPlaneNodes); i++ {
			randomString, err := utils.GenerateRandomString(8)
			if err != nil {
				panic(err)
			}
			var mainControlPlaneNode types.ServerCreateRequest
			mainControlPlaneNode.Server.Name = kaasRequest.Name + "-cp-" + randomString
			mainControlPlaneNode.Server.Flavor = "3"
			mainControlPlaneNode.Server.KeyName = kaasRequest.KeyName
			mainControlPlaneNode.Server.Image = "e62a8d62-02f9-4bbd-ae13-ca298faec579" // image : k8s-centos8-openstack
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
			if *kaasRequest.ControlPlaneNodes[i].Main {
				members = append(members, types.Node{
					FixedIP: kaasRequest.ControlPlaneNodes[i].FixedIP,
					Name:    randomString,
					Kind:    "cp",
					Main:    kaasRequest.ControlPlaneNodes[i].Main,
				})
				mainControlPlaneNode.Server.Networks = []types.ServerCreateNetwork{
					{
						UUID:    kaasRequest.NetworkId,
						FixedIP: &kaasRequest.ControlPlaneNodes[i].FixedIP,
					},
					{
						// UUID is manually assigned. But It will be different in another openstack environment.
						// mgmt : "abe6e177-dd45-4923-9cbd-ae2a09ce4fb9", provider :"d64dbfc6-46de-4f2c-8b65-12c9d29a8b7e",
						UUID:    "abe6e177-dd45-4923-9cbd-ae2a09ce4fb9",
						FixedIP: nil,
					},
				}
				mainControlPlaneNode.Server.Script = encodedScript
				responseOfServerCreation, err := openstack.CreateServer(tokenID, projectID, mainControlPlaneNode)
				if err != nil {
					panic(err)
				}
				fmt.Println("Main : ", mainControlPlaneNode.Server.Name, responseOfServerCreation)
				go func() {
					mainResultChannel <- responseOfServerCreation
				}()
			} else {
				members = append(members, types.Node{
					FixedIP: kaasRequest.ControlPlaneNodes[i].FixedIP,
					Name:    randomString,
					Kind:    "cp",
				})
				mainControlPlaneNode.Server.Networks = []types.ServerCreateNetwork{
					{
						UUID:    kaasRequest.NetworkId,
						FixedIP: &kaasRequest.ControlPlaneNodes[i].FixedIP,
					},
				}

				responseOfServerCreation, err := openstack.CreateServer(tokenID, projectID, mainControlPlaneNode)
				if err != nil {
					panic(err)
				}
				fmt.Println(mainControlPlaneNode.Server.Name, responseOfServerCreation)
				go func(i int) {
					resultChannel <- kaasRequest.ControlPlaneNodes[i].FixedIP
				}(i)
			}
		}

		// 4. Create the Data-plane nodes on openstack (without shell script)
		// 	  And send the result to the resultChannel
		fmt.Println("Create the Data-plane nodes on openstack (without shell script)")
		for i := 0; i < len(kaasRequest.DataPlaneNodes); i++ {
			randomString, err := utils.GenerateRandomString(8)
			if err != nil {
				panic(err)
			}
			members = append(members, types.Node{
				FixedIP: kaasRequest.DataPlaneNodes[i].FixedIP,
				Name:    randomString,
				Kind:    "dp",
			})
			var mainControlPlaneNode types.ServerCreateRequest
			mainControlPlaneNode.Server.Name = kaasRequest.Name + "-dp-" + randomString
			mainControlPlaneNode.Server.Flavor = "3"
			mainControlPlaneNode.Server.KeyName = kaasRequest.KeyName
			mainControlPlaneNode.Server.Image = "e62a8d62-02f9-4bbd-ae13-ca298faec579"
			mainControlPlaneNode.Server.SecurityGroups = []types.ServerCreateSecurityGroup{
				{
					Name: "default",
				},
				{
					Name: "k8s-common",
				},
				{
					Name: "k8s-data-plane",
				},
			} // image : k8s-centos8-openstack
			mainControlPlaneNode.Server.Networks = []types.ServerCreateNetwork{
				{
					UUID:    kaasRequest.NetworkId,
					FixedIP: &kaasRequest.DataPlaneNodes[i].FixedIP,
				},
				// {
				// 	// UUID is manually assigned. But It will be different in another openstack environment.
				// 	// mgmt : "abe6e177-dd45-4923-9cbd-ae2a09ce4fb9", provider :"d64dbfc6-46de-4f2c-8b65-12c9d29a8b7e",
				// 	UUID:    "d64dbfc6-46de-4f2c-8b65-12c9d29a8b7e",
				// 	FixedIP: nil,
				// },
			}
			responseOfServerCreation, err := openstack.CreateServer(tokenID, projectID, mainControlPlaneNode)
			if err != nil {
				panic(err)
			}
			fmt.Println(mainControlPlaneNode.Server.Name, responseOfServerCreation)
			go func(i int) {
				resultChannel <- kaasRequest.DataPlaneNodes[i].FixedIP
			}(i)
		}

		// 5. Handle creation response from main control-plane node and run SSH command to get join command
		mainResponce := <-mainResultChannel
		server, ok := mainResponce["server"].(map[string]interface{})
		if !ok {
			fmt.Println("Invalid server data")
			panic(ok)
		}
		id, ok := server["id"].(string)
		if !ok {
			fmt.Println("Invalid id")
			panic(ok)
		}
		// 5.1 get fixedMgmtIP and status of main control-plane node If status is ACTIVE, get fixedMgmtIP
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
				time.Sleep(5 * time.Second)
			}
		}

		// 5.2 run SSH command with some timeout to get kubeadm join command
		// Private key file has to be saved in the same directory as the main.go file (not implimented yet)
		keyFile := filepath.Join(filepath.Dir(mainPath), "ssh/k8s.pem")
		sshClient := utils.SSH{
			IP:   fixedMgmtIP, // use Mgmt IP, but it is not known yet.
			User: "centos",
			Port: 22,
			Cert: keyFile, // Password or Key file path
		}
		var kubeadmJoinOutput string
		timeout = time.After(120 * time.Second) // Set the timeout to 120 seconds
		remainingTime := 120 * time.Second
	OuterLoopForSSH:
		for kubeadmJoinOutput == "" {
			select {
			case <-timeout:
				fmt.Println("Timeout reached")
				break OuterLoopForSSH // Exit the outer loop if timeout is reached
			default:
				fmt.Println("Remaining time For SSH connection try to Main Control Plane:", remainingTime)
				kubeadmJoinOutput, err = utils.GetKubeadmJoinOutput(sshClient, utils.CertPublicKeyFile)
				if err != nil {
					time.Sleep(10 * time.Second)
					remainingTime -= 10 * time.Second
					break
				}
				break OuterLoopForSSH
			}
		}

		// 6. Make to join k8s cluster for Data-plane nodes by giving and running shell script
		// 	  From Main control-plane node to ready-Data-plane nodes
		var wg sync.WaitGroup
		wg.Add(len(kaasRequest.DataPlaneNodes) + len(kaasRequest.ControlPlaneNodes) - 1)
		go func() {
			for i := 0; i < len(kaasRequest.DataPlaneNodes)+len(kaasRequest.ControlPlaneNodes)-1; i++ {
				defer func() {
					wg.Done()
					if r := recover(); r != nil {
						// Handle the panic here, such as logging the error
						fmt.Println("Panic occurred:", r)
					}
				}()
				createdIP := <-resultChannel
				// 6.1 SSH to Data-plane node through Main control-plane node as proxy. And run kubeadm join command
				_, err := utils.InjectDataplaneJoin(sshClient, utils.CertPublicKeyFile, `sudo `+kubeadmJoinOutput, createdIP)
				if err != nil {
					panic(err)
				}
			}
		}()
		wg.Wait()

		// 7. Save Nodes information(cluster info) in DB (MongoDB)
		k8sCluster.Members = members
		utils.MongoDBInsertCluster(mongoDB, k8sCluster)

		c.JSON(http.StatusOK, gin.H{
			"keypair": keypair,
			"cmd":     kubeadmJoinOutput,
			"members": members,
		})
	}
}
