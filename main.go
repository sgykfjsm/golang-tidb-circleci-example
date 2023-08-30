package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sgykfjsm/golang-tidb-circleci-example/manager"
	"github.com/sgykfjsm/golang-tidb-circleci-example/mydb"
	"github.com/sgykfjsm/golang-tidb-circleci-example/util"
)

var username, password, host, port, dbName string

func LogFatalIfNotNil(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	username = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	host = os.Getenv("DB_HOST")
	port = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_NAME")

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("[MyApp] ")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port)
	log.Printf("Going to open the connection with dsn %s", dsn)
	db, err := sql.Open("mysql", dsn)
	LogFatalIfNotNil(err)

	args := os.Args
	if len(args) < 2 {
		log.Fatal("Need to pass SQL file name to initialize the database")
	}
	fileName := args[1]
	ctx := context.Background()

	log.Print("Start initializing the database")
	err = initializeDB(db, ctx, fileName)
	LogFatalIfNotNil(err)
}

func initializeDB(db *sql.DB, ctx context.Context, fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var queries []string

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		queries = util.ParseQueryLine(text, queries)
	}

	for _, query := range queries {
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var err error

	log.Printf("DBName: %s", dbName)
	config := mysql.Config{
		User:                 username,
		Passwd:               password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", host, port),
		DBName:               dbName,
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
	}

	dsn := config.FormatDSN()
	log.Printf("Going to open the connection with dsn %s", dsn)

	db, err := sql.Open("mysql", dsn)
	LogFatalIfNotNil(err)
	defer func(db *sql.DB, dbName string) {
		cleanupDatabase(db, dbName)
		db.Close()
	}(db, dbName)

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(1 * time.Second)

	ctx := context.Background()
	userManager := manager.NewUserManager(db, ctx)

	user := mydb.User{Name: "John Doe"}

	log.Println("Insert data")
	lastID, err := userManager.AddUser(user)
	LogFatalIfNotNil(err)

	res, err := userManager.GetUserByID(lastID)
	LogFatalIfNotNil(err)
	log.Printf("Get user ID: %d, Name: %s\n", res.ID, res.Name)

	newUser := mydb.User{
		ID:   lastID,
		Name: "Jane Doe",
	}

	log.Printf("Update user %#v)", newUser)
	err = userManager.UpdateUser(newUser)
	LogFatalIfNotNil(err)

	anotherUser := mydb.User{Name: "foo bar baz"}
	log.Println("Insert data again")
	_, err = userManager.AddUser(anotherUser)
	LogFatalIfNotNil(err)

	log.Println("List all users")
	users, err := userManager.ListUsers()
	LogFatalIfNotNil(err)
	for i, user := range users {
		log.Printf("%d: user ID: %d, Name: %s\n", i, user.ID, user.Name)
	}

	log.Println("end")
}

func cleanupDatabase(db *sql.DB, dbName string) {
	db.Exec("DROP DATABASE IF EXISTS " + dbName)
}
