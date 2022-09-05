[English](README.md) | 🇨🇳中文
# jsonrpc4go
## 🧰 安装
```
go get -u github.com/sunquakes/jsonrpc4go
```
## 📖 开始使用
- 服务端代码
```go
package main

import (
    "github.com/sunquakes/jsonrpc4go"
)

type IntRpc struct{}

type Params struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Result = int

func (i *IntRpc) Add(params *Params, result *Result) error {
	a := params.A + params.B
	*result = interface{}(a).(Result)
	return nil
}

func main() {
	s, _ := jsonrpc4go.NewServer("http", "127.0.0.1", "3232") // http协议
	s.Register(new(IntRpc))
	s.Start()
}
```
- 客户端代码
```go
package main

import (
	"fmt"
	"github.com/sunquakes/jsonrpc4go"
)

type Params struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Result = int

type Result2 struct {
	C int `json:"c"`
}

func main() {
	result := new(Result)
	c, _ := jsonrpc4go.NewClient("http", "127.0.0.1", "3232") // http协议
	err := c.Call("IntRpc/Add", Params{1, 6}, result, false) // 路由支持以下3种格式: "int_rpc/Add", "int_rpc.Add", "IntRpc.Add"
	// 发送的数据格式: {"id":"1604283212", "jsonrpc":"2.0", "method":"IntRpc/Add", "params":{"a":1,"b":6}}
	// 接收的数据格式: {"id":"1604283212", "jsonrpc":"2.0", "result":7}
	fmt.Println(err) // nil
	fmt.Println(*result) // 7
}
```
## ⚔️ 测试
```
go test -v ./test/...
```
## 🚀 更多特性
- tcp协议
```go
s, _ := jsonrpc4go.NewServer("tcp", "127.0.0.1", "3232") // tcp协议

c, _ := jsonrpc4go.NewClient("tcp", "127.0.0.1", "3232") // tcp协议
```
- 钩子 (在代码's.Start()'前添加下面的代码)
```go
// 在方法前执行的钩子方法
s.SetBeforeFunc(func(id interface{}, method string, params interface{}) error {
    // 如果方法返回error类型，服务端停止执行并返回错误信息到客户端
    // 例：return errors.New("Custom Error")
    return nil
})
// 在方法后执行的钩子方法
s.SetAfterFunc(func(id interface{}, method string, result interface{}) error {
    // 如果方法返回error类型，服务端停止执行并返回错误信息到客户端
    // 例：return errors.New("Custom Error")
    return nil
})
```
- 限流 (在代码's.Start()'前添加下面的代码)
```go
s.SetRateLimit(20, 10) // 最大并发数为10, 最大请求数为每秒20个
```
- tcp协议时自定义请求结束符
```go
// 在代码's.Start()'前添加下面的代码
s.SetOptions(server.TcpOptions{"aaaaaa", nil}) // 仅tcp协议生效
// 在代码'c.Call()'或'c.BatchCall()'前添加下面的代码
c.SetOptions(client.TcpOptions{"aaaaaa", nil}) // 仅tcp协议生效
```
- 通知请求
```go
// 通知
result2 := new(Result2)
err2 := c.Call("int_rpc/Add2", Params{1, 6}, result2, true) // 第一个参数可以是"IntRpc/Add2"、"int_rpc.Add2"或"IntRpc.Add2"
// 发送的数据格式: {"jsonrpc":"2.0","method":"IntRpc/Add2","params":{"a":1,"b":6}}
// 接收的数据格式: {"jsonrpc":"2.0","result":{"c":7}}
fmt.Println(err2) // nil
fmt.Println(*result2) // {7}
```
- 批量请求
```go
// 批量请求
result3 := new(Result)
err3 := c.BatchAppend("IntRpc/Add1", Params{1, 6}, result3, false)
result4 := new(Result)
err4 := c.BatchAppend("IntRpc/Add", Params{2, 3}, result4, false)
c.BatchCall()
// 发送的数据格式: [{"id":"1604283212","jsonrpc":"2.0","method":"IntRpc/Add1","params":{"a":1,"b":6}},{"id":"1604283212","jsonrpc":"2.0","method":"IntRpc/Add","params":{"a":2,"b":3}}]
// 接收的数据格式: [{"id":"1604283212","jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found","data":null}},{"id":"1604283212","jsonrpc":"2.0","result":5}]
fmt.Println((*err3).Error()) // Method not found
fmt.Println(*result3) // 0
fmt.Println(*err4) // nil
fmt.Println(*result4) // 5
```
## 📄 License
`jsonrpc4go`代码遵守[Apache-2.0 license](/LICENSE)开源协议。
