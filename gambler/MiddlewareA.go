package gambler

import (
	"log"
	"time"
)

func MiddlewareA() HandlerFunc {
	return func(c *Context) {
		t := time.Now()
		log.Printf("Debug msg : MiddlewareA.go -> MiddlewareA : START middle ware [ MiddlewareA ]\n")
		c.Next()
		log.Printf("Debug msg : MiddlewareA.go -> MiddlewareA : END middleware [ MiddlewareA ] with msg = [%d] %s in %v\n", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
