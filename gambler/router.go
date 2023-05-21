package gambler

import (
	"log"
	"net/http"
	"strings"
)

//router.go:将路由相关的方法提取出来，方便后续更新路由有关的功能

// router 定义路由的存储方式
// roots key eg, roots['GET'] roots['POST']
// handlers key eg, handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']
type router struct {
	roots    map[string]*node       // 存储每种请求方式的 Trie 树根节点
	handlers map[string]HandlerFunc // 存储每个路由对应的 HandlerFunc
}

// NewRouter 提供路由实例的创建函数
func NewRouter() *router {
	log.Printf("Debug msg : router.go -> NewRouter : create router with roots, handlers\n")
	log.Printf("CREATE NewRouter FINISH\n")
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

//  parsePattern 解析 pattern, 得到每一个小 part, 保存到 parts 中
func parsePattern(pattern string) []string {
	// 将 pattern 也就是完整的路径 以 / 进行分割
	vs := strings.Split(pattern, "/")
	// 用于保存分割后的每个 part
	//log.Printf("Debug msg : router.go -> parsePattern : vs = %v\n", vs)
	parts := make([]string, 0)
	for _, part := range vs {
		//  如果这个 part 不是空则加入到 parts 中
		if part != "" {
			parts = append(parts, part)
			// part 如果以 * 开头，则是模糊匹配，后面的 part 不需要保存了
			if part[0] == '*' {
				break
			}
		}
	}
	return parts
}

// addRouter 功能是添加路由，也就是添加前缀树的节点
func (r *router) addRouter(method string, pattern string, handler HandlerFunc) {
	log.Printf("Debug msg : router.go -> addRouter : method = %v, pattern = %v\n", method, pattern)
	_, ok := r.roots[method]
	// 如果该方法还没有 tire 树则创建
	if !ok {
		r.roots[method] = &node{}
	}
	// 将该节点插入到 tire 树中，并设置 handler
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	log.Printf("Debug msg : router.go -> addRouter : parsePattern res is : parts = %v, key = %v\n", parts, key)
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

// getRoute 功能是查找路，得到解析结果和前缀树中对应的节点
// 解析了 : 和 * 两种匹配符的参数，返回一个 map
// eg: searchParts:/p/go/doc, nodeParts:/p/:lang/doc, 解析结果为：{lang: "go"}
// eg: static/css/geektutu.css匹配到/static/*filepath，解析结果为{filepath: "css/indexpage.css"}
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	log.Printf("Debug msg : router.go -> getRoute : method = %v, path = %v\n", method, path)
	params := make(map[string]string)
	// 尝试得到对应请求方式的 前缀树根节点
	root, ok := r.roots[method]
	log.Printf("Debug msg : router.go -> getRoute :get method root = %v\n", root)
	if !ok {
		return nil, nil
	}
	// 解析请求路径得到 parts
	searchParts := parsePattern(path)
	log.Printf("Debug msg : router.go -> getRoute : get searchParts from path = %v\n", searchParts)
	// 成功拿到对应请求的前缀树后查找匹配当前路径的节点
	n := root.search(searchParts, 0)
	log.Printf("Debug msg : router.go -> getRoute : get node = %v\n", n)
	if n != nil {
		// 解析这个节点的路径
		parts := parsePattern(n.pattern)
		log.Printf("Debug msg : router.go -> getRoute : get node parts = %v\n", parts)
		// 准备解析的结果，用 params 保存，key 是 节点， value 是 路径
		// 利用 index 实现对应
		for index, part := range parts {
			// :开头的只需要拿到这个 part
			if part[0] == ':' {
				// part[1:] = lang, searchParts[index] = go
				params[part[1:]] = searchParts[index]
			}
			// *开头的使用 [index:] 一次性拿到了路径
			if part[0] == '*' && len(part) > 1 {
				// part[1:] = filepath, 后面是把剩余的 part 用 / 连接
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

// getRoutes 返回的nodes就是一个个已注册的route，比如/p/:lang/hello 等
// 用于测试
func (r *router) getRoutes(method string) []*node {
	// 拿到某一种请求方法的 前缀树 所有节点
	root, ok := r.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	// 调用travel这个方法来获取该Method作为root下的所有route
	root.travel(&nodes)
	return nodes
}

// handle 的参数改为 context
func (r *router) handle(c *Context) {
	// 拿到前缀树的节点
	n, params := r.getRoute(c.Method, c.Path)
	log.Printf("Debug msg : router.go -> handle : node = %v\n", n)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		// r.handlers[key] 是和当前路由对应的 handlerFunc
		// 这一步骤是将与这个路由匹配的 handler 函数添加到 handlers 列表中
		// 这个列表中已经包含了要执行的中间件，是在前一步 ServeHTTP 中添加的
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND : %s\n", c.Path)
		})
	}
	log.Printf("Debug msg : router.go -> handle : nums of handlers = %d\n", len(c.handlers))
	// Next() 中从上下文的 handlers 列表中拿出中间件和 handler 执行
	c.Next()
}

// showTree 展示某一路径的节点 node
func (r *router) showTree(method string, path string) {
	root, ok := r.roots[method]
	log.Printf("Debug msg : router.go -> showTree : show root = %v\n", root)
	if !ok {
		log.Printf("Debug msg : router.go -> showTree : PATH NOT FOUND\n")
		return
	}
	searchParts := parsePattern(path)
	// 成功拿到对应请求的前缀树后查找匹配当前路径的节
	node := root.search(searchParts, 0)
	log.Printf("Debug msg : router.go -> showTree : node struct :\n\t node.pattern = %v\n\t node.part = %v\n\t node.children = %v\n\t node.isWild = %v\n\t", node.pattern, node.part, node.children, node.isWild)
}
