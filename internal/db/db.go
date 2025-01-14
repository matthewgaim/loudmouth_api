package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func ConnectToDatabase() (*sql.DB, error) {
	fmt.Println("Connecting to DB")

	pg_username := os.Getenv("PG_USERNAME")
	pg_password := os.Getenv("PG_PASSWORD")
	pg_db_ip := os.Getenv("PG_DB_IP")
	pg_db_name := os.Getenv("PG_DB_NAME")

	connStr := fmt.Sprintf("postgresql://%v:%v@%v/%v?sslmode=disable", pg_username, pg_password, pg_db_ip, pg_db_name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	fmt.Println("Connected to DB!")
	return db, nil
}

func ConnectToRedis() *redis.Client {
	redis_address := os.Getenv("REDIS_ADDRESS")
	redis_port := os.Getenv("REDIS_PORT")
	addr := fmt.Sprintf("%v:%v", redis_address, redis_port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
		Protocol: 2,
	})
	return client
}
