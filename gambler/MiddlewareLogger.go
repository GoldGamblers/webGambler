package gambler

import (
	"log"
	"time"
)

func MiddlewareLogger() HandlerFunc {
	return func(c *Context) {
		t := time.Now()
		log.Printf("Debug msg : MiddlewareLogger.go -> MiddlewareLogger : START middle ware [ MiddlewareLogger ]\n")
		c.Next()
		log.Printf("Debug msg : MiddlewareLogger.go -> logger : END middle ware [ MiddlewareLogger ] with msg = [%d] %s in %v\n\n", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
