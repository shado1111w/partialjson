# Partial Json - Go 语言 JSON 补全解析

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/shado1111w/partialjson.svg)](https://pkg.go.dev/github.com/shado1111w/partialjson)
[![Go Report Card](https://goreportcard.com/badge/github.com/shado1111w/partialjson)](https://goreportcard.com/report/github.com/shado1111w/partialjson)
[![codecov](https://codecov.io/gh/shado1111w/partialjson/graph/badge.svg?token=5B0NBMQTGS)](https://codecov.io/gh/shado1111w/partialjson)

很多时候，我们想尽快处理大模型流式返回的 JSON 数据，而不是等到所有数据都返回后才进行解析。标准库(`encoding/json`)和其他第三方 JSON 库（如`jsoniter`,`sonic`,`gjson`等）都要求输入的 JSON 数据是完整的，否则会报错。这里提供了一种解决方案，可以在不完整的 JSON 数据上进行补全或者解析。

## ✨ 功能特性

- 🛠 **自动补全** - 修复不完整的 JSON 数据
- 📈 **高效性能** - 基于 Go 的高效解析实现

## 📦 安装

```bash
go get github.com/shado1111w/partialjson
```

## 🚀 快速开始
### 基本用法
#### 补全 JSON 数据
```go
package main

import (
    "fmt"
    "github.com/shado1111w/partialjson"
)

func main() {
    input := `{"name": "Alice", "age": 30, "hobbies": ["reading", "coding"`
    
	// strictMode: 是否启用严格模式
	// 如果启用，会忽略不完整的字符串值（这在语义上是不完整的），例如：{"name": "Alic 的输出结果为{"name": null}
	// 如果不启用，将不考虑字符串值语义是否完整，直接保存下来，例如：{"name": "Alic 的输出结果为{"name": "Alic"}
	parser := partialjson.NewJSONParser(true)
	result, err := parser.EnsureJson(input)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("完整JSON结构:\n%s", result)
    // 输出:
    // {
    //   "name": "Alice",
    //   "age": 30,
    //   "hobbies": ["reading", "coding"]
    // }
}
```
#### 解析 JSON 数据
```go
package main

import (
    "fmt"
    "github.com/shado1111w/partialjson"
)

func main() {
    input := `{"name": "Alice", "age": 30, "hobbies": ["reading", "coding"`
    
	parser := partialjson.NewJSONParser(true)
	var data map[string]interface{}
	err := parser.Unmarshal([]byte(input), &data)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("解析结果:\n%v", data)
    // 输出:
    // map[age:30 hobbies:[reading coding] name:Alice]
}
```
### 进阶
#### 快速补全 JSON 数据（在流式解析中比EnsureJson快将近1倍）
```go
package main

import (
    "fmt"
    "github.com/shado1111w/partialjson"
)

func main() {
    input := `{"name": "Alice", "age": 30, "hobbies": ["reading", "coding"`
    
	parser := partialjson.NewJSONParser(true)
	result, err := parser.FastEnsureJson(input)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("完整JSON结构:\n%s", result)
    // 输出:
    // {
    //   "name": "Alice",
    //   "age": 30,
    //   "hobbies": ["reading", "coding"]
    // }
}
```

## 📚 应用场景
- 💬 大模型流式 JSON 输出解析
