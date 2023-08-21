package manager_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sgykfjsm/golang-tidb-circleci-example/manager"
	"github.com/sgykfjsm/golang-tidb-circleci-example/mydb"
	"golang.org/x/exp/slices"
	"gotest.tools/assert"
)

var createDatabaseDDL = `
CREATE DATABASE IF NOT EXISTS %s;
USE %s;
CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL AUTO_RANDOM PRIMARY KEY,
    name VARCHAR(64) NOT NULL
);`
var username, password, host, port string

func init() {
	username = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	host = os.Getenv("DB_HOST")
	port = os.Getenv("DB_PORT")
}

func setUp(queries []string) (*sql.DB, context.Context, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?multiStatements=true", username, password, host, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return nil, nil, err
		}
	}

	return db, ctx, nil
}

func TestAddUser(t *testing.T) {
	dbName := "test_addUser"

	queries := []string{
		fmt.Sprintf(createDatabaseDDL, dbName, dbName),
	}
	db, ctx, err := setUp(queries)
	if err != nil {
		t.Fatalf("failed to setup the connection to the database: %s, err: %s", dbName, err)
	}
	defer func(dbName string, db *sql.DB) {
		db.Exec("DROP DATABASE IF EXISTS " + dbName)
		db.Close()
	}(dbName, db)

	newUser := mydb.User{Name: "addtest_user"}
	userManager := manager.NewUserManager(db, ctx)
	id, err := userManager.AddUser(newUser)
	if err != nil {
		t.Fatalf("failed to execute main.AddUser err: %s", err)
	}

	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		t.Fatalf("failed to get user by ID %d err: %s", id, err)
	}

	assert.Equal(t, newUser.Name, name)
}

func TestUpdateUser(t *testing.T) {
	dbName := "test_updateUser"

	queries := []string{
		fmt.Sprintf(createDatabaseDDL, dbName, dbName),
		`INSERT INTO users (name) VALUES ("test_user")`,
	}
	db, ctx, err := setUp(queries)
	if err != nil {
		t.Fatalf("failed to setup the connection to the database: %s, err: %s", dbName, err)
	}
	defer func(dbName string, db *sql.DB) {
		db.Exec("DROP DATABASE IF EXISTS " + dbName)
		db.Close()
	}(dbName, db)

	var id int64
	err = db.QueryRowContext(ctx, "SELECT id FROM users LIMIT 1").Scan(&id)
	if err != nil {
		t.Fatalf("failed to get user err: %s", err)
	}

	userManager := manager.NewUserManager(db, ctx)
	newName := "new_user"
	newUser := mydb.User{
		ID:   id,
		Name: newName,
	}

	if err = userManager.UpdateUser(newUser); err != nil {
		t.Fatalf("failed to execute main.AddUser err: %s", err)
	}

	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		t.Fatalf("failed to get user by ID %d err: %s", id, err)
	}

	assert.Equal(t, newUser.Name, name)
}

func TestGetUserByID(t *testing.T) {
	dbName := "test_selectUserByID"
	testUserName := "test_user"

	queries := []string{
		fmt.Sprintf(createDatabaseDDL, dbName, dbName),
		fmt.Sprintf("INSERT INTO users (`name`) VALUES (%q)", testUserName),
	}
	db, ctx, err := setUp(queries)
	if err != nil {
		t.Fatalf("failed to setup the connection to the database: %s, err: %s", dbName, err)
	}
	defer func(dbName string, db *sql.DB) {
		db.Exec("DROP DATABASE IF EXISTS " + dbName)
		db.Close()
	}(dbName, db)

	var id int64
	err = db.QueryRowContext(ctx, "SELECT id FROM users LIMIT 1").Scan(&id)
	if err != nil {
		t.Fatalf("failed to get user err: %s", err)
	}

	userManager := manager.NewUserManager(db, ctx)

	user, err := userManager.GetUserByID(id)
	if err != nil {
		t.Fatalf("failed to get user by ID %d err: %s", id, err)
	}

	assert.Equal(t, user.Name, testUserName)
}

func TestListUsers(t *testing.T) {
	dbName := "test_selectUserByID"
	testUserNames := []string{
		"test_user001",
		"test_user002",
	}

	queries := []string{
		fmt.Sprintf(createDatabaseDDL, dbName, dbName),
		fmt.Sprintf("INSERT INTO users (`name`) VALUES (%q)", testUserNames[0]),
		fmt.Sprintf("INSERT INTO users (`name`) VALUES (%q)", testUserNames[1]),
	}
	db, ctx, err := setUp(queries)
	if err != nil {
		t.Fatalf("failed to setup the connection to the database: %s, err: %s", dbName, err)
	}
	defer func(dbName string, db *sql.DB) {
		db.Exec("DROP DATABASE IF EXISTS " + dbName)
		db.Close()
	}(dbName, db)

	userManager := manager.NewUserManager(db, ctx)
	users, err := userManager.ListUsers()
	if err != nil {
		t.Fatalf("failed to list users err: %s", err)
	}

	assert.Equal(t, len(testUserNames), len(users))
	for _, user := range users {
		assert.Assert(t, slices.Index(testUserNames, user.Name) > -1)
	}
}
