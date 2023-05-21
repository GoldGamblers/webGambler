package gambler

import (
	"log"
	"time"
)

func MiddlewareB() HandlerFunc {
	return func(c *Context) {
		t := time.Now()
		log.Printf("Debug msg : MiddlewareB.go -> MiddlewareB : START middle ware [ MiddlewareB ]\n")
		c.Next()
		log.Printf("Debug msg : MiddlewareB.go -> MiddlewareB : END middleware [ MiddlewareB ] with msg = [%d] %s in %v\n", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
