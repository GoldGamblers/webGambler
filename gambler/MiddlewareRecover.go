package gambler

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

//用来获取触发 panic 的堆栈信息
func trace(message string) string {
	var pcs [32]uintptr
	// Callers 用来返回调用栈的程序计数器, 第 0 个 Caller 是 Callers 本身，第 1 个是上一层 trace，第 2 个是再上一层的 defer func
	// 为了日志简洁跳过了前 3 个 Caller。
	n := runtime.Callers(3, pcs[:])
	var str strings.Builder
	str.WriteString(message + "\n\t\t\t\t\tTraceback:")
	for _, pc := range pcs[:n] {
		// 获取对应的函数
		fn := runtime.FuncForPC(pc)
		// 获取到调用该函数的文件名和行号，打印在日志中
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t\t\t\t\t\t%s:%d", file, line))
	}
	return str.String()
}

// MiddlewareRecover 错误恢复的处理函数
func MiddlewareRecover() HandlerFunc {
	return func(c *Context) {
		log.Printf("Debug msg : MiddlewareRecover.go -> MiddlewareRecover : START middle ware [ MiddlewareRecover ]\n")
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("Debug msg : MiddlewareRecover.go -> MiddlewareRecover : panic!  %s \n", trace(message))
				log.Printf("Debug msg : MiddlewareRecover.go -> MiddlewareRecover : RECOVER SUCCESS\n")
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		// 必须要有这个next，这代表接下来执行其他中间件和用户的handler(接来来是什么取决于中间件调用的顺序)
		c.Next()
		// 如果没有这个就无法 recover 到用户 handler，如果发生了错误，会转到 defer去执行，所以下面的这一行log只会在没有错误时正常输出
		log.Printf("Debug msg : MiddlewareRecover.go -> MiddlewareRecover : END middle ware [ MiddlewareRecover ]\n")
	}
}
