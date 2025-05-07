package internal

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
)

var (
	rateLock sync.Mutex
	clients  *cache.Cache
)

func init() {
	clients = cache.New(time.Second, 15*time.Second)
}

func RateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ip, _, _ := net.SplitHostPort(c.Request().RemoteAddr)

		rateLock.Lock()
		defer rateLock.Unlock()
		_, exists := clients.Get(ip)
		if exists {
			return c.JSON(http.StatusTooManyRequests, "Too many requests")
		}
		clients.Set(ip, time.Now(), cache.DefaultExpiration)

		return next(c)
	}
}
