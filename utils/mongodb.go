package utils

import (
	"context"
	"fmt"
	"kaps/types"
)

type Cluster struct {
	Name   string `json:"name"`
	Member []Node `json:"member"`
}

type Node struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func MongoDBInsertCluster(mongoDB *types.MongoDB, cluster types.K8SCluster) {

	client := mongoDB.Client

	coll := client.Database("kaps").Collection("cluster")

	result, err := coll.InsertOne(context.TODO(), cluster)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
}
