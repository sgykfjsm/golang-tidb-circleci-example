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
	"github.com/sgykfjsm/golang-tidb-circleci-example/mydb"
)

var username, password, host, port, dbName string

func LogFatalIfNotNil(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func AddUser(db *sql.DB, ctx context.Context, user mydb.User) (int64, error) {
	q := mydb.New(db)
	res, err := q.AddUser(ctx, user.Name)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func UpdateUser(db *sql.DB, ctx context.Context, user mydb.User) error {
	q := mydb.New(db)

	return q.UpdateUser(ctx, mydb.UpdateUserParams{
		ID:   user.ID,
		Name: user.Name,
	})
}

func GetUserByID(db *sql.DB, ctx context.Context, id int64) (mydb.User, error) {
	q := mydb.New(db)

	return q.GetUserByID(ctx, id)
}

func ListUsers(db *sql.DB, ctx context.Context) ([]mydb.User, error) {
	q := mydb.New(db)

	return q.ListUsers(ctx)
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

func parseQueryLine(text string, queries []string) []string {
	if strings.HasPrefix(text, "--") || len(queries) == 0 {
		queries = append(queries, text)
		return queries
	}

	lines := strings.Split(text, ";")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if i == 0 && len(queries) > 0 && !strings.HasSuffix(queries[len(queries)-1], ";") {
			queries[len(queries)-1] += " " + line
		} else {
			queries = append(queries, line)
		}

		if i != len(lines)-1 {
			queries[len(queries)-1] += ";"
		}
	}

	return queries
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
		queries = parseQueryLine(text, queries)
	}

	for _, query := range queries {
		// log.Printf("Exec query: %s", query)
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
	user := mydb.User{Name: "John Doe"}

	log.Println("Insert data")
	lastID, err := AddUser(db, ctx, user)
	LogFatalIfNotNil(err)

	res, err := GetUserByID(db, ctx, lastID)
	LogFatalIfNotNil(err)
	log.Printf("Get user ID: %d, Name: %s\n", res.ID, res.Name)

	newUser := mydb.User{
		ID:   lastID,
		Name: "Jane Doe",
	}

	log.Printf("Update user %#v)", newUser)
	err = UpdateUser(db, ctx, newUser)
	LogFatalIfNotNil(err)

	anotherUser := mydb.User{Name: "foo bar baz"}
	log.Println("Insert data again")
	_, err = AddUser(db, ctx, anotherUser)
	LogFatalIfNotNil(err)

	log.Println("List all users")
	users, err := ListUsers(db, ctx)
	LogFatalIfNotNil(err)
	for i, user := range users {
		log.Printf("%d: user ID: %d, Name: %s\n", i, user.ID, user.Name)
	}

	log.Println("end")
}

func cleanupDatabase(db *sql.DB, dbName string) {
	db.Exec("DROP DATABASE IF EXISTS " + dbName)
}
