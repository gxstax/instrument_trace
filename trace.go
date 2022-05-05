package trace

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

func printTrace(id uint64, name, arrow string, indent int)  {
	indentsContent := ""
	for i := 0; i < indent; i++ {
		indentsContent += "    "
	}
	fmt.Printf("g[%05d]:%s%s%s\n", id, indentsContent, arrow, name)
}

// 保存每个 goroutine 缩进值
// key: goroutineID, value: 缩进层次
var m = make(map[uint64]int)
// map 不支持并发写，增加 mu 同步对 map 的写操作
var mu sync.Mutex

var goroutineSpace = []byte("goroutine ")

func curGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]

	// Parse the 4707 out of "goroutine 4707 ["
	b = bytes.TrimPrefix(b, goroutineSpace)
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		panic(fmt.Sprintf("No space found in %q", b))
	}
	b = b[:i]
	n, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse goroutine ID out of %q: %v", b, err))
	}
	return n
}

// Trace 跟踪函数名的自动获取 (添加 goroutineID)
func Trace() func() {
	/*
		Caller函数四个返回值:
			pc uintptr: 程序计数器
			file string: 所在源文件
			line int: 所在文件行
			ok bool: 是否成功获取到信息
	*/
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("not found caller")
	}

	//
	fn := runtime.FuncForPC(pc)
	name := fn.Name()
	gid := curGoroutineID()

	mu.Lock()
	indents := m[gid]
	m[gid] = indents + 1
	mu.Unlock()
	printTrace(gid, name, "->", indents)
	return func() {
		mu.Lock()
		indents := m[gid]
		m[gid] = indents - 1
		mu.Unlock()
		printTrace(gid, name, "<-", indents - 1)
	}
}