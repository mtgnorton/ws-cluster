package main

import "github.com/gogf/gf/v2/frame/g"

// 静态文件服务器基本使用
func main() {
	s := g.Server()
	s.SetServerRoot("./index")
	s.SetPort(8333)
	s.Run()
}
