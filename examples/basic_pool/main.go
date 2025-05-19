package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/shrimps80/go-service-utils/pool"
)

// 计数器
var counter int32

func main() {
	// 创建一个大小为5的协程池
	p, err := pool.New(5)
	if err != nil {
		log.Fatalf("创建协程池失败: %v", err)
	}

	// 提交20个任务
	for i := 1; i <= 20; i++ {
		taskID := i
		err := p.Submit(func() error {
			return processTask(taskID)
		})
		if err != nil {
			log.Printf("提交任务 %d 失败: %v", taskID, err)
		}
	}

	// 等待所有任务完成
	p.Wait()

	// 获取统计信息
	stats := p.Stats()
	fmt.Printf("协程池统计: 运行任务=%d, 等待任务=%d, 已完成任务=%d\n",
		stats.RunningTasks, stats.WaitingTasks, stats.CompletedTasks)

	// 关闭协程池
	p.Close()

	fmt.Printf("所有任务已完成，总计处理: %d\n", atomic.LoadInt32(&counter))
}

// processTask 模拟处理任务
func processTask(id int) error {
	fmt.Printf("开始处理任务 %d\n", id)

	// 随机工作时间，模拟不同任务的处理时间
	workTime := time.Duration(rand.Intn(3)+1) * time.Second
	time.Sleep(workTime)

	// 增加计数器
	atomic.AddInt32(&counter, 1)

	fmt.Printf("任务 %d 已完成，耗时: %v\n", id, workTime)
	return nil
}
