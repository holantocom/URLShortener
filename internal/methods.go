package internal

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/jxskiss/base62"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
)

type URL struct {
	Short    string `json:"short"`
	Original string `json:"original"`
	Clicks   int    `json:"clicks"`
}

var urlCache *cache.Cache

func init() {
	urlCache = cache.New(15*time.Second, 30*time.Second)
}

func Shorten(c echo.Context) error {
	var u URL
	if err := json.NewDecoder(c.Request().Body).Decode(&u); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid input")
	}

	if !isValidURL(u.Original) {
		return c.JSON(http.StatusBadRequest, "Invalid URL format")
	}

	if u.Original == "" || len(u.Original) > 2048 {
		return c.JSON(http.StatusBadRequest, "URL required and must be <2048 chars")
	}

	var lastId int64
	db := c.Get("db").(*sql.DB)
	err := db.QueryRow("INSERT INTO urls (original, clicks) VALUES ($1, 0) RETURNING id", u.Original).Scan(&lastId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "DB error")
	}
	u.Short = string(base62.FormatInt(lastId))

	return c.JSON(http.StatusOK, u)
}

func Redirect(c echo.Context) error {
	short := c.Param("short")

	URLId, err := base62.ParseInt([]byte(short))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Error while parsing short ID")
	}

	var u URL
	db := c.Get("db").(*sql.DB)
	cachedURL, exists := urlCache.Get(short)
	if exists {
		u.Original = cachedURL.(string)
	} else {
		if err = db.QueryRow("SELECT original FROM urls WHERE id = $1", URLId).Scan(&u.Original); err != nil {
			return c.JSON(http.StatusNotFound, "URL not found")
		}

		urlCache.Set(short, u.Original, cache.DefaultExpiration)
	}

	go func() {
		if _, err := db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE id = $1", URLId); err != nil {
			log.Println("Failed to increment clicks:", err)
		}
	}()

	// use 302 "Moved Temporarily" because HTTP 301 use browser cache
	return c.Redirect(http.StatusFound, u.Original)
}

func Stats(c echo.Context) error {
	short := c.Param("short")

	URLId, err := base62.ParseInt([]byte(short))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Error while parsing short ID")
	}

	var u URL
	db := c.Get("db").(*sql.DB)
	err = db.QueryRow("SELECT original, clicks FROM urls WHERE id = $1", URLId).Scan(&u.Original, &u.Clicks)
	if err != nil {
		return c.JSON(http.StatusNotFound, "URL not found")
	}
	u.Short = short

	return c.JSON(http.StatusOK, u)
}

func isValidURL(raw string) bool {
	res, err := url.ParseRequestURI(raw)
	return err == nil && res.Hostname() != ""
}
