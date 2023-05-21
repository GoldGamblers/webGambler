package main

import (
	"example/tools"
	"fmt"
	"gambler"
	"html/template"
	"net/http"
	"time"
)

// test URL
//url: http://localhost:9999/
//url: http://localhost:9999/hello
//url: http://localhost:9999/hello?name=liup2
//url: http://localhost:9999/hello/liup2
//url: http://localhost:9999/hello/liup2/doc
//url: http://localhost:9999/assets/file.txt

//url: http://localhost:9999/g1
//url: http://localhost:9999/g1/hello
//url: http://localhost:9999/g1/hello?name=liup2

//url: http://localhost:9999/g2/
//url: http://localhost:9999/g2/hello/liup2
//url: http://localhost:9999/g2/assets/file.txt

//url: http://localhost:9999/date
//url: http://localhost:9999/students

type student struct {
	Name string
	Age  int8
}

func main() {
	r := gambler.New()
	// 添加自定义的全局中间件 MiddlewareLogger 和 MiddlewareRecover
	r.UseMiddlewares(gambler.MiddlewareLogger(), gambler.MiddlewareRecover())

	// 测试recover中间件
	r.GET("/panic", func(c *gambler.Context) {
		names := []string{"liup2"}
		c.String(http.StatusOK, names[10])
	})

	//测试模板能否正常加载和渲染
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": tools.FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")
	studentA := &student{Name: "Liu", Age: 20}
	studentB := &student{Name: "Zhou", Age: 18}
	r.GET("/", func(c *gambler.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(c *gambler.Context) {
		c.HTML(http.StatusOK, "showStudents.tmpl", gambler.JsonMap{
			"title":  c.Query("name"),
			"stuArr": [2]*student{studentA, studentB},
		})
	})
	r.GET("/date", func(c *gambler.Context) {
		c.HTML(http.StatusOK, "showTime.tmpl", gambler.JsonMap{
			"title": "gambler",
			"now":   time.Now(),
		})
	})

	// 路由注册测试
	r.GET("/hello", func(c *gambler.Context) {
		// /hello?name=liup2
		c.String(http.StatusOK, "Hello %s, here is %s\n", c.Query("name"), c.Path)
	})

	r.GET("/hello/:name", func(c *gambler.Context) {
		// expect /hello/liup2
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.GetParam("name"), c.Path)
	})

	r.GET("/hello/:name/doc", func(c *gambler.Context) {
		// expect /hello/liup2/doc
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.GetParam("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *gambler.Context) {
		c.JSON(http.StatusOK, gambler.JsonMap{"filepath": c.GetParam("filepath")})
	})

	r.POST("/login", func(c *gambler.Context) {
		c.JSON(http.StatusOK, gambler.JsonMap{
			"userName": c.PostForm("userName"),
			"passWord": c.PostForm("passWord"),
		})
	})

	r.PUT("/put", func(c *gambler.Context) {
		fmt.Fprintf(c.Writer, "Method: PUT")
	})

	// 分组测试 g1
	g1 := r.NewGroup("/g1")
	// 测试只给 g1 分组添加中间件
	g1.UseMiddlewares(gambler.MiddlewareA())
	{
		g1.GET("/", func(c *gambler.Context) {
			c.HTML(http.StatusOK, "showGroupIndex.tmpl", gambler.JsonMap{"group": "g1"})
		})

		g1.GET("/hello", func(c *gambler.Context) {
			// /g1/hello?name=liup2
			//c.String(http.StatusOK, "Hello %s, here is %s\n", c.PostForm("name"), c.Path)
			c.String(http.StatusOK, "Hello %s, here is %s\n", c.Query("name"), c.Path)
		})
	}

	// 分组测试 g2
	g2 := r.NewGroup("/g2")
	// 测试只给 g2 分组添加中间件
	g2.UseMiddlewares(gambler.MiddlewareB())
	{
		g2.GET("/", func(c *gambler.Context) {
			c.HTML(http.StatusOK, "showGroupIndex.tmpl", gambler.JsonMap{"group": "g2"})
		})

		g2.GET("/hello/:name", func(c *gambler.Context) {
			// /g1/hello/liup2
			c.String(http.StatusOK, "Hello %s, here is %s\n", c.GetParam("name"), c.Path)
		})

		g2.GET("/assets/:filepath", func(c *gambler.Context) {
			c.JSON(http.StatusOK, gambler.JsonMap{
				"filepath": c.GetParam("filepath"),
			})
		})

		g2.POST("/login", func(c *gambler.Context) {
			c.JSON(http.StatusOK, gambler.JsonMap{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
	}

	//r.ShowTree("GET", "/hello")
	err := r.Run(":9999")
	if err != nil {
		return
	}
}
