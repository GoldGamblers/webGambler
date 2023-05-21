package gambler

import (
	"fmt"
	"log"
	"strings"
)

// tireTree.go: 用前缀树来匹配路由，实现动态路由功能
// 包括 参数匹配 和 通配*
// Tips: 前缀树的路由必须匹配到叶子节点才可以，不可以中途下车
// 比如注册了 /hello/doc，访问 /hello，会存在 part 为 hello 的节点，但是 pattern 为空，就无法匹配到，返回404

// 前缀树节点
type node struct {
	pattern  string  // 准备匹配的路由，eg: /p/:lang
	part     string  // 路由中的一部分，eg: :lang
	children []*node // 子节点，eg: [doc, tutorial, intro]
	isWild   bool    // 是否是模糊匹配，part含有 : 或 * 时为Ture
}

// matchChild 第一个匹配成功的节点，用于更新前缀树，设置新的路由
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			log.Printf("Debug msg : tireTree.go -> matchChild : part = %s, child = %v\n", part, child)
			return child
		}
	}
	log.Printf("Debug msg : tireTree.go -> matchChild : match child with part { %s } failed, do insert\n", part)
	return nil
}

// matchChildren 查找所有匹配的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			//fmt.Printf("Debug msg : tireTree.go -> matchChildren : part = %s, child = %v\n", part, child)
			nodes = append(nodes, child)
		}
	}
	log.Printf("Debug msg : tireTree.go -> matchChildren : part = %s, nodes = %v\n", part, nodes)
	return nodes
}

// insert 插入节点
func (n *node) insert(pattern string, parts []string, height int) {
	log.Printf("Debug msg : tireTree.go -> insert : pattern = %s, parts = %v, height = %v\n", pattern, parts, height)
	if len(parts) == height {
		// 遍历完了所有的part，那就把路径写到这个节点的pattern字段中
		n.pattern = pattern
		log.Printf("Debug msg : tireTree.go -> insert : insert node = %v FINISH\n", n)
		return
	}
	//log.Printf("Debug msg : tireTree.go -> insert : parts = %v\n", parts)
	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		log.Printf("Debug msg : tireTree.go -> insert : child = %v", child)
		n.children = append(n.children, child)
	}
	// 递归 height + 1
	child.insert(pattern, parts, height+1)
}

// search 查找节点
func (n *node) search(parts []string, height int) *node {
	log.Printf("Debug msg : tireTree.go -> search : parts = %v\n", parts)
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		// 如果是只注册了 /hello/doc 这样的路径，那么当想访问 /hello 时，就会找不到 /hello 对应的节点，因为 /hello 路径对应节点的 pattern 为空
		if n.pattern == "" {
			log.Printf("Debug msg : tireTree.go -> search : node with part{ %s } found, but node pattern is null, NOT REGISTER\n", n.part)
			return nil
		}
		return n
	}
	part := parts[height]
	//log.Printf("Debug msg : tireTree.go -> search : part = %v, index = %v\n", part, height)
	children := n.matchChildren(part)
	//log.Printf("Debug msg : tireTree.go -> search : children = %v\n", children)
	for _, child := range children {
		// 递归 height + 1
		result := child.search(parts, height+1)
		if result != nil {
			log.Printf("Debug msg : tireTree.go -> search : result = %v\n", result)
			return result
		}
	}
	log.Printf("Debug msg : tireTree.go -> SEARCH FAILED \n")
	return nil
}

// travel 查找所有完整的url，保存到列表中
// 用于测试
func (n *node) travel(list *([]*node)) {
	if n.pattern != "" {
		// 递归终止条件
		*list = append(*list, n)
	}
	// 一层一层的递归找pattern是非空的节点
	for _, child := range n.children {
		child.travel(list)
	}
}

// String 打印节点值
func (n *node) String() string {
	return fmt.Sprintf("node{ pattern=%s, part=%s, isWild=%t }", n.pattern, n.part, n.isWild)
}
