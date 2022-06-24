# [课题]Job-Scheduler

### 快速开始
```
$ go get github.com/invxp/job-scheduler
```
### 这是一个单机的任务调度器
1. 支持动态新增任务
2. 支持间隔性的重复执行任务
3. 支持数据持久化(via PebbleDB)
4. 对于执行失败(超时)的任务会进行自动重试
5. 单机理论支持百万个任务并发执行

### 涉及的知识点
1. Gob数据结构序列与反序列化(用于持久保存)
2. PebbleDB的基础使用(LSMBased-KVStore)
3. Context与Channel的使用技巧
4. 一开始想用时间轮来调度的,后来想了一下,还是用协程吧

### 从这里开始吧!!!
```go
package main

import (
	"github.com/invxp/job-scheduler/pkg/scheduler"
	"log"
	"time"
)

func main() {
	s, e := scheduler.New("test")
	if e != nil {
		panic(e)
	}

	t, e := s.Add("MyTask", time.Second*5)

	if e != nil {
		panic(e)
	}

	log.Println("new task scheduled", t)

	s.Run()
}

```

#### 测试用例可以这样做

```
$ go test -v -race -run @XXXXX(具体方法名)
PASS / FAILED
```

#### 或测试全部用例:
```
$ go test -v -race
```

## 注意
* 不要用在生产环境,这只是个玩具
