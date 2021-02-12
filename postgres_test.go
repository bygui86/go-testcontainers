package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	// WARN: really important otherwise "database/sql" is not able to find the "postgres" driver and test fails!
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	password = "secret"

	productName  = "sample"
	productPrice = 9.90

	createTableQuery = `CREATE TABLE IF NOT EXISTS products(
		id SERIAL,
		name TEXT NOT NULL,
		price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
		CONSTRAINT products_pkey PRIMARY KEY (id)
	)`
	insertQuery = "INSERT INTO products(name, price) VALUES($1, $2) RETURNING id"
	selectQuery = "SELECT id,name,price FROM products ORDER BY id ASC LIMIT $1 OFFSET $2"
	deleteQuery = "DELETE FROM products WHERE id = $1"
)

type product struct {
	ID    int
	Name  string
	Price float64
}

func TestPostgres(t *testing.T) {
	ctx := context.Background()

	postgres, pgErr := startPostgres(ctx)
	require.NoError(t, pgErr)
	fmt.Println("PostgreSQL up and running")

	host, port, hostPortErr := getHostAndPort(postgres, ctx)
	assert.NotEqual(t, "", host)
	assert.NotEqual(t, "", port)
	require.NoError(t, hostPortErr)

	sqlClient, clientErr := createSqlClient(host, port)
	require.NoError(t, clientErr)

	pingErr := pingPostgres(sqlClient)
	require.NoError(t, pingErr)

	_, tableErr := createTable(sqlClient)
	require.NoError(t, tableErr)

	_, checkErr := checkTable(sqlClient)
	require.NoError(t, checkErr)

	product := &product{
		Name:  productName,
		Price: productPrice,
	}

	insertErr := insertRow(sqlClient, ctx, product)
	require.NoError(t, insertErr)

	products, selectErr := selectRows(sqlClient, ctx)
	require.NoError(t, selectErr)
	assert.NotNil(t, products)
	assert.NotEmpty(t, products)
	assert.Len(t, products, 1)
	assert.NotEqual(t, 0, products[0].ID)
	assert.Equal(t, productName, products[0].Name)
	assert.Equal(t, productPrice, products[0].Price)

	_, deleteErr := deleteRow(sqlClient, ctx, products[0].ID)
	require.NoError(t, deleteErr)

	// INFO: can be replaced by `defer postgres.Terminate(ctx)`
	exitErr := postgres.Terminate(ctx)
	if exitErr != nil {
		fmt.Printf("PostgreSQL termination failed: %s\n", exitErr.Error())
	}
}

func startPostgres(ctx context.Context) (testcontainers.Container, error) {
	fmt.Println("Starting PostgreSQL container")

	contReq := testcontainers.ContainerRequest{
		Image:        "postgres:13.1-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": password,
			// "POSTGRES_HOST_AUTH_METHOD":       "trust",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	return testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: contReq,
			Started:          true,
		},
	)
}

func getHostAndPort(postgres testcontainers.Container, ctx context.Context) (string, nat.Port, error) {
	expPorts, expErr := postgres.Ports(ctx)
	if expErr != nil {
		return "", "", expErr
	}

	fmt.Println("PostgreSQL exposed ports:")
	for k, v := range expPorts {
		fmt.Printf("\t %s -> %v \n", k, v)
	}

	host, hostErr := postgres.Host(ctx)
	if hostErr != nil {
		return "", "", hostErr
	}

	port, portErr := postgres.MappedPort(ctx, "5432")
	if portErr != nil {
		return "", "", portErr
	}

	fmt.Printf("PostgreSQL host: %s \n", host)
	fmt.Printf("PostgreSQL port: %d \n", port.Int())
	return host, port, nil
}

func createSqlClient(host string, port nat.Port) (*sql.DB, error) {
	fmt.Println("Create SQL client")

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port.Int(),
		"postgres", password, "postgres",
		"disable",
	)
	return sql.Open("postgres", connString)
}

func pingPostgres(sqlClient *sql.DB) error {
	fmt.Println("Ping")
	return backoff.Retry(
		func() error {
			err := sqlClient.Ping()
			if err != nil {
				log.Println("PostgreSQL not ready, backing off...")
				return err
			}
			log.Println("PostgreSQL ready for further test")
			return nil
		},
		backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 10),
	)
}

func createTable(sqlClient *sql.DB) (sql.Result, error) {
	fmt.Println("Create table")
	return sqlClient.Exec(createTableQuery)
}

func checkTable(sqlClient *sql.DB) (sql.Result, error) {
	fmt.Println("Check table")
	return sqlClient.Exec("SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'products'")
}

func insertRow(sqlClient *sql.DB, ctx context.Context, prod *product) error {
	fmt.Println("Insert row")

	return sqlClient.
		QueryRowContext(
			ctx,
			insertQuery,
			prod.Name,
			prod.Price,
		).
		Scan(&prod.ID)
}

func selectRows(sqlClient *sql.DB, ctx context.Context) ([]*product, error) {
	fmt.Println("Select rows")

	rows, queryErr := sqlClient.QueryContext(ctx, selectQuery, 10, 0)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	products := make([]*product, 0)
	for rows.Next() {
		var prod product
		rowErr := rows.Scan(&prod.ID, &prod.Name, &prod.Price)
		if rowErr != nil {
			return nil, rowErr
		}
		products = append(products, &prod)
	}
	return products, nil
}

func deleteRow(sqlClient *sql.DB, ctx context.Context, productId int) (sql.Result, error) {
	fmt.Println("Delete row")

	return sqlClient.ExecContext(ctx, deleteQuery, productId)
}
