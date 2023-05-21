package gambler

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

//gambler.go: 网络框架入口

// HandlerFunc 定义handler，处理函数,使用 Context 封装 req 和 resp
type HandlerFunc func(c *Context)

type RouterGroup struct {
	prefix      string        // 例如 / 或者 /api
	middlewares []HandlerFunc // 支持中间件
	parent      *RouterGroup  // 为了支持嵌套分组，需要知道父分组
	engine      *Engine       // 需要有访问 router 的能力，所以保存一个指向 engine 的指针，方便通过 engine 访问各种接口，也意味着框架的资源由 engine 协调
}

// Engine 定义实例引擎,集中保存管理路由
// *RouterGroup 是一个嵌套类型，是指将已有的类型直接声明在新的结构类型里。
// *RouterGroup 被称作内部类型 Engine 被称为外部类型。
// 内部类型的属性、方法，可以为外部类型所有，就好像是外部类型自己的一样。
// 外部类型还可以定义自己的属性和方法，甚至可以定义与内部相同的方法，这样内部类型的方法就会被“屏蔽”
type Engine struct {
	router        *router            // 定义路由：key 是理由，value 是处理函数
	*RouterGroup                     // engine 是最顶层的分组，拥有 RouterGroup 的所有能力
	groups        []*RouterGroup     // 保存所有的 group
	htmlTemplates *template.Template // 使用 html/template 的渲染能力，把模板加载到内存中(还有一个text/template)
	funcMap       template.FuncMap   // 保存所有的自定义模板渲染函数, 是一个map
}

// New 构造函数
func New() *Engine {
	log.Printf("Debug msg : gambler.go -> New : create web engine with router, RouterGroup, groups\n")
	// 实例化 engine 的 路由对象
	engine := &Engine{router: NewRouter()}
	// 实例化 engine 的 分组对象，表示分组对象可以通过engine访问一些接口
	engine.RouterGroup = &RouterGroup{engine: engine}
	// 实例化 engine 的 groups 对象，存放多个RouterGroup分组
	engine.groups = []*RouterGroup{engine.RouterGroup}
	log.Printf("ENGINE CREATE FINISH\n\n")
	return engine
}

// ServeHTTP 实现 ServeHTTP，需要分组，因为所有的路径都需要 ServeHTTP
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 为适配中间件而增添的部分
	var middlewares []HandlerFunc
	// 拿到和请求对应的分组的所有 中间件 并赋值给 上下文的 hanslers 列表
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	// 实例化一个 Context
	c := newContext(w, req)
	// 将中间件列表添加到这个上下文的 hanslers 列表中
	c.handlers = middlewares
	// 用于 Context 使用 engine 的方法
	c.engine = engine
	log.Printf("Debug msg : gambler.go -> ServeHTTP : create context finish\n")
	engine.router.handle(c)
}

// Run 封装监听函数，监听函数不需要分组，因为所有的路径都需要监听
func (engine *Engine) Run(addr string) (err error) {
	log.Printf("LISTENING ADDR = %s\n\n", addr)
	return http.ListenAndServe(addr, engine)
}

// ShowTree 打印某种请求方式的路由前缀树
func (engine *Engine) ShowTree(method string, path string) {
	// 竟然可以套娃
	//engine.RouterGroup.engine.router.showTree(method, path)
	engine.router.showTree(method, path)
}

// SetFuncMap 用于设置自定义函数渲染模板 funcMap
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
	log.Printf("Debug msg : gambler.go -> SetFuncMap : SET funcMap FINISH\n")
}

// LoadHTMLGlob 用于加载模板
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// NewGroup 创建 RouterGroup, 所有的 groups 都共享一个相同的 engine 接口
func (group *RouterGroup) NewGroup(prefix string) *RouterGroup {
	log.Printf("Debug msg : gambler.go -> NewGroup : create NewGroup with prefix = %v, newGroup prefix = %v\n", prefix, group.prefix+prefix)
	engine := group.engine
	// 新的分组
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	// 保存新创建的分组
	engine.groups = append(engine.groups, newGroup)
	log.Printf("GROUP REGISTER FINISH: group = %s\n\n", newGroup.prefix)
	return newGroup
}

// addRoute 实现添加路由功能：method是请求方式，pattern是路径，handlerFunc是处理函数
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	log.Printf("Debug msg : gambler.go -> addRoute : method = %s, pattern = %s + %s\n", method, group.prefix, comp)
	// comp 是不包含前缀的路径，在真正添加路由的时候需要拼接起来。
	// 如果没有调用新建分组那么这个前缀会设置为空
	pattern := group.prefix + comp
	//log.Printf("***** comp = %s *****", comp)
	// router.addRouter 需要通过 engine 来调用
	group.engine.router.addRouter(method, pattern, handler)
	log.Printf("ROUTE REGISTER FINISH : method = %s, path = %s \n\n", method, pattern)
}

// GET 实现 GET 路由：pattern是路径，handlerFunc是处理函数
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	log.Printf("Debug msg : gambler.go -> GET : pattern = %s\n", pattern)
	// addRoute 不需要通过 engine 来调用
	group.addRoute("GET", pattern, handler)
	//log.Printf("ROUTE REGISTER FINISH : method = GET, path = %s \n\n", pattern)
}

// POST 实现 POST 路由：pattern是路径，handlerFunc是处理函数
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	log.Printf("Debug msg : gambler.go -> POST : pattern = %s\n", pattern)
	// addRoute 不需要通过 engine 来调用
	group.addRoute("POST", pattern, handler)
	//log.Printf("ROUTE REGISTER FINISH : method = POST, path = %s \n\n", pattern)
}

// PUT 实现 PUT 路由：pattern是路径，handlerFunc是处理函数
func (group *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	log.Printf("Debug msg : gambler.go -> PUT : pattern = %s\n", pattern)
	// addRoute 不需要通过 engine 来调用
	group.addRoute("PUT", pattern, handler)
	//log.Printf("ROUTE REGISTER FINISH : method = PUT, path = %s \n\n", pattern)
}

// UseMiddlewares 将中间件应用到某一个 group 中
func (group *RouterGroup) UseMiddlewares(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
	log.Printf("Debug msg : gambler.go -> UseMiddleWare : use middlewares:%v for group : %v\n", group.middlewares, group.prefix)
}

// createStaticHandler 创建静态文件的 handler，浏览器收到html文件，会自动执行加载css，发起http请求
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	// 拿到绝对路径
	absolutePath := path.Join(group.prefix, relativePath)
	// StripPrefix将URL中的前缀中的prefix字符串删除，然后再交给后面的Handler处理，返回值是一个 handler
	// http.FileServer 返回一个 handler，
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	log.Printf("Debug msg : gambler.go -> createStaticHandler : absolutePath = %v + %v, FINISH STATIC FILE HANDLER\n", group.prefix, relativePath)
	return func(c *Context) {
		// 获取文件名
		file := c.GetParam("filepath")
		if _, err := fs.Open(file); err != nil {
			// 文件打开失败
			c.SetStatus(http.StatusNotFound)
			return
		}
		// 拿到文件后就可以交给 http.FileServer 来处理了
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static 用于映射路径，可以将磁盘上某个文件夹的 root 映射到 relativePath
func (group *RouterGroup) Static(relativePath string, root string) {
	log.Printf("Debug msg : gambler.go -> Static : register static file\n")
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}
