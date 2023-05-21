package gambler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// context.go: 封装*http.Request和http.ResponseWriter的方法，简化相关接口的调用，只是设计 Context 的原因之一
// 对于框架来说，还需要支撑额外的功能，如解析动态路由、支持中间件等
// 对于中间件来说，需要支持用户自定义功能插入到框架中，因为框架无法理解全部的业务逻辑。需要考虑 插入点 和 中间件的输入
// 中间件一般要求能在 handler 之前和之后都执行一些操作，因此需要一个切换函数

// JsonMap 给用于保存JSON数据的map起一个别名
type JsonMap map[string]interface{}

// Context 构建上下文的字段
type Context struct {
	// 原始字段
	Writer     http.ResponseWriter
	Req        *http.Request
	Path       string            // req 请求信息
	Method     string            // req 请求信息
	StatusCode int               // resp 响应信息
	Params     map[string]string // 保存解析后的参数
	handlers   []HandlerFunc     // 中间件部分：这个列表中表示里面的 handler 可能会结合中间件进行处理
	index      int               // 中间件部分：表示执行到了第几个中间件
	engine     *Engine           // 用于能够通过 Context 来访问 engine 的 HTML 模板，在实例化的时候需要给 engine 赋值
}

// newContext 创建新的 context
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	log.Printf("Debug msg : context.go -> newContext : create context\n")
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

// Next 用于切换中间件，在中间件调用该方法时将会把控制权交给下一个中间件，直到最后一个中间件，然后在从后往前调用每个中间件在 next 之后的部分
func (c *Context) Next() {
	c.index++
	num := len(c.handlers)
	for ; c.index < num; c.index++ {
		log.Printf("Debug msg : context.go -> Next : change next in []HandlerFunc, index = %d, num of handlers = %d\n", c.index, num)
		// 切换中间件的控制，参考匿名函数的执行
		// 匿名函数的最后传入参数就代表执行
		c.handlers[c.index](c)
	}
}

// PostForm Tip：net/http包下 Request.FormValue 方法 可以额获取 url 中? 后面的请求参数，或者是已解析的表单数据
// PostForm 封装 FromValue 方法,获取表单中指定 key 的值
func (c *Context) PostForm(key string) string {
	log.Printf("Debug msg : context.go -> PostForm : key = %s, value = %s \n", key, c.Req.FormValue(key))
	return c.Req.FormValue(key)
}

// Query Tip：Req.URL.Query().Get(key) 可以额获取 url 中? 后面的请求参数，一般是 GET 方法常用
// Query 通过 key 获取 URL 中的对应的查询参数值
func (c *Context) Query(key string) string {
	log.Printf("Debug msg : context.go -> Query : key = %s, value = %s \n", key, c.Req.URL.Query().Get(key))
	return c.Req.URL.Query().Get(key)
}

// SetStatus 设置响应头，也就是状态码
func (c *Context) SetStatus(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
	log.Printf("Debug msg : context.go -> SetStatus : code = %v\n", c.StatusCode)
}

// SetHeader 设置 Header, 例子如下
// Header["Connection"] = ["keep-alive"]
// Header["User-Agent"] = ["Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"]
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
	log.Printf("Debug msg : context.go -> SetHeader : key = %s, value = %s \n", key, value)
}

// String 构造 string 类型响应的方法，其中 ... 表示可以接受任意数量的接口类型
func (c *Context) String(code int, format string, value ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	// 设置响应头的响应码
	c.SetStatus(code)
	// ... 的用法
	// 第一个用法主要是用于函数有多个不定参数的情况，表示为可变参数，可以接受任意个数但相同类型的参数
	// 第二个用法是slice可以被打散进行传递
	c.Writer.Write([]byte(fmt.Sprintf(format, value...)))
	log.Printf("Debug msg : context.go -> String : write content = %s", []byte(fmt.Sprintf(format, value...)))
}

// JSON 构造 JSON 类型响应的方法，接口类型可以表示任意值
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content_type", "application/json")
	c.SetStatus(code)
	encoder := json.NewEncoder(c.Writer)
	log.Printf("Debug msg : context.go -> JSON : finish\n")
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// Data 构造 Data 类型响应的方法，接口类型可以表示任意值
func (c *Context) Data(code int, data []byte) {
	c.SetStatus(code)
	c.Writer.Write(data)
	log.Printf("Debug msg : context.go -> Data : data = %v\n", data)
}

// HTML 构造 HTML 类型的响应方法，接口类型可以表示任意值, 可以根据模板文件名选择模板进行渲染
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.SetStatus(code)
	log.Printf("Debug msg : context.go -> HTML : template name = %s\n", name)
	//c.Writer.Write([]byte(name)) // 被 ExecuteTemplate 代替
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
}

// GetParam 提供获取到url中 key 对应的值的方法
func (c *Context) GetParam(key string) string {
	value, _ := c.Params[key]
	log.Printf("Debug msg : context.go -> GetParam : value = %v\n", value)
	return value
}

// Fail 错误信息反馈
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, JsonMap{"message": err})
}
