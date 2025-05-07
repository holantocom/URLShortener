package main

import (
	"URLShortener/internal"

	"database/sql"
	"log"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	dbRepo := internal.NewPostgresURLRepository(db)

	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(10 * time.Minute)

	e := echo.New()
	e.Use(middleware.Recover())
	// can be replaced with e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))
	e.Use(internal.RateLimiter)

	e.POST("/shorten", internal.Shorten(dbRepo))
	e.GET("/:short", internal.Redirect(dbRepo))
	e.GET("/stats/:short", internal.Stats(dbRepo))

	e.Logger.Fatal(e.Start(":8080"))
}
