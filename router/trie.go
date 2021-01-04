package router

import (
	"errors"
	"strings"
)

const (
	paramSeparator    = ":"
	filePathSeparator = "*"
)

// Trie树节点
type node struct {
	pattern  string
	part     string
	children []*node
	isWild   bool
}

// /test/:name/hello
// 插入匹配 只找到一个合适的路由
func (n *node) matchOne(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 查询匹配 找到多个合适的路由 例如对于/test/doc可匹配出/test/:lang和/test/doc和/test/:xxx
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// pattern 总路径
// part 按照/拆分的子路径
// height 递归层数
func (n *node) insert(pattern string, parts []string, height int) error {
	// 递归终止条件 递归到底 同时也表示路径添加过 覆盖
	if len(parts) == height {
		n.pattern = pattern
		return nil
	}
	// 获取当前匹配的节点
	part := parts[height]
	err, isWild := pathValid(part)
	if err != nil {
		return err
	}
	// 寻找符合的路径 若没有则创建 继续递归直到parts遍历完成
	child := n.matchOne(part)
	if child == nil {
		child = &node{
			part:     part,
			children: make([]*node, 0),
			isWild:   isWild != "",
		}
		n.children = append(n.children, child)
	}
	//
	return child.insert(pattern, parts, height+1)
}
func pathValid(part string) (error, string) {
	paramName := ""
	// 判断是否是参数路径
	if strings.Contains(part, paramSeparator) {
		split := strings.Split(part, paramSeparator)
		if len(split) > 2 {
			// 参数不合法
			return errors.New("路径中只能包含一个:用作通配符"), paramName
		}
		paramName = split[1]
	}
	// 判断是否是参数路径
	if strings.HasPrefix(part, filePathSeparator) {
		split := strings.Split(part, filePathSeparator)
		if len(split) > 2 {
			// 参数不合法
			return errors.New("路径中只能包含一个*用作通配符"), paramName
		}
		paramName = split[1]
	}
	return nil, paramName
}

// 请求路径寻找 解析正确的handler
func (n *node) search(pattern string, parts []string, height int) *node {
	// 判断若是*开头的或者已经递归到底了 *开头表示文件 不需要
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == pattern {
			return n
		}
		if n.part == "" {
			return nil
		}
		return n
	}
	// 拿到当前节点的路径
	part := parts[height]
	children := n.matchChildren(part)
	var c *node
	// 遍历子节点的路径 看有没有符合的
	for _, child := range children {
		if child.pattern == pattern {
			return child
		}
		c = child.search(pattern, parts, height+1)
	}
	return c
}
