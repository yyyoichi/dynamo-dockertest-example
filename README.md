# dynamo-dockertest-example

Example of DynamoDB Local test with ory/dockertest

It is required to schema to be created before running the tests.

The assumption in this example is to migrate the schema before running the test,
and expected to clean up the items after the test is done.

## Example

```golang

// main_test.go

var DynamoDBClient *dynamodb.Client

func TestMain(m *testing.M) {

    const port = "32808"
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

```

## Container Manegements

In this example, the Container is running unless you explicitly stop it.
This is because it takes a relatively long time to connect to the container (this is not limited to DynamoDB).

At least in the local environment, you should be able to run more tests smoothly.

## Expansion

If parallel testing is required, it would be better to cut out the code, launch a container for each port, and delete the tables in that container after the test.

```golang

func newDynamoClient(t *testing.T, port string) *dynamodb.Client {
    // Same code as TestMain ...
    return client
}

func testDropTable(t *testing.T, client *dynamodb.Client) {
    //
}

```

Alternatively, it may be better to devise a migration and use different tables between tests, and delete the tables.

```golang
func testMigrate(t *testing.T, client *dynamodb.Client, tableName string) {
    // make table and indexes
}

```
