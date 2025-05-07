package internal

import (
	"context"
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

func Shorten(repo URLRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		lastId, err := repo.Save(c.Request().Context(), u.Original)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "DB error")
		}
		u.Short = string(base62.FormatInt(lastId))

		return c.JSON(http.StatusOK, u)
	}
}

func Redirect(repo URLRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		short := c.Param("short")

		URLId, err := base62.ParseInt([]byte(short))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Error while parsing short ID")
		}

		var u URL
		cachedURL, exists := urlCache.Get(short)
		if exists {
			u.Original = cachedURL.(string)
		} else {
			data, _ := repo.FindByID(c.Request().Context(), URLId)
			if data == nil {
				return c.JSON(http.StatusNotFound, "URL not found")
			}

			u.Original = data.Original
			urlCache.Set(short, u.Original, cache.DefaultExpiration)
		}

		go func() {
			if err = repo.IncrementClicks(context.Background(), URLId); err != nil {
				log.Println("Failed to increment clicks:", err)
			}
		}()

		// use 302 "Moved Temporarily" because HTTP 301 use browser cache
		return c.Redirect(http.StatusFound, u.Original)
	}
}

func Stats(repo URLRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		short := c.Param("short")

		URLId, err := base62.ParseInt([]byte(short))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Error while parsing short ID")
		}

		data, _ := repo.FindByID(c.Request().Context(), URLId)
		if data == nil {
			return c.JSON(http.StatusNotFound, "URL not found")
		}
		data.Short = short

		return c.JSON(http.StatusOK, data)
	}
}

func isValidURL(raw string) bool {
	res, err := url.ParseRequestURI(raw)
	return err == nil && res.Hostname() != ""
}
