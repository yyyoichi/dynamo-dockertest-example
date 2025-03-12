package main_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var DynamoDBClient *dynamodb.Client

func TestMain(m *testing.M) {

	var port = "32808"
	var ContainerName = fmt.Sprintf("dynamodb-local-%s", port)

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	if resource, found := pool.ContainerByName(ContainerName); !found {
		_, err = pool.RunWithOptions(
			&dockertest.RunOptions{
				Name:       ContainerName,
				Repository: "amazon/dynamodb-local",
				Tag:        "latest",
				// -inMemory flag is used to run dynamodb in memory
				// if stopped, all data will be lost
				Cmd: []string{"-jar", "DynamoDBLocal.jar", "-inMemory", "-port", "8000"},
				PortBindings: map[docker.Port][]docker.PortBinding{
					"8000/tcp": {{HostPort: port}},
				},
			},
		)
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}
	} else {
		// existing container
		// 1. get container status and check if it is running
		if !resource.Container.State.Running {
			// 2. if not running, start the container
			err = pool.Client.StartContainer(resource.Container.ID, nil)
			if err != nil {
				log.Fatalf("Could not start resource: %s", err)
			}
			// 3. migrate schema if you need to
		}
	}

	endpoint := fmt.Sprintf("http://localhost:%s", port)
	DynamoDBClient = dynamodb.New(dynamodb.Options{
		BaseEndpoint: &endpoint,
		Credentials:  credentials.NewStaticCredentialsProvider("dummy", "dummy", "dummy")})

	if err := pool.Retry(func() error {
		_, err := DynamoDBClient.ListTables(context.Background(), &dynamodb.ListTablesInput{})
		fmt.Println("Retrying connection: ", err)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to DynamoDB: %s", err)
	}

	m.Run()
}

func TestListTable(t *testing.T) {
	_, err := DynamoDBClient.ListTables(context.Background(), &dynamodb.ListTablesInput{})
	if err != nil {
		t.Errorf("Error listing tables: %v", err)
	}
}
