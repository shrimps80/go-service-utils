package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/shrimps80/go-service-utils/pool"
)

func main() {
	// 创建一个大小为3的带优先级的协程池
	p, err := pool.New(3, pool.WithPriority())
	if err != nil {
		log.Fatalf("创建协程池失败: %v", err)
	}

	// 随机种子
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 提交15个任务，随机分配优先级
	priorities := []pool.Priority{
		pool.PriorityLow,
		pool.PriorityNormal,
		pool.PriorityHigh,
	}

	priorityNames := map[pool.Priority]string{
		pool.PriorityLow:    "低",
		pool.PriorityNormal: "中",
		pool.PriorityHigh:   "高",
	}

	for i := 1; i <= 15; i++ {
		taskID := i
		priority := priorities[rand.Intn(len(priorities))]

		err := p.SubmitWithOptions(func() error {
			return processTask(taskID, priorityNames[priority])
		}, pool.WithTaskPriority(priority))

		if err != nil {
			log.Printf("提交任务 %d (优先级: %s) 失败: %v", taskID, priorityNames[priority], err)
			continue
		}

		fmt.Printf("提交任务 %d (优先级: %s)\n", taskID, priorityNames[priority])
	}

	// 等待所有任务完成
	p.Wait()

	// 获取统计信息
	stats := p.Stats()
	fmt.Printf("协程池统计: 运行任务=%d, 等待任务=%d, 已完成任务=%d\n",
		stats.RunningTasks, stats.WaitingTasks, stats.CompletedTasks)

	// 关闭协程池
	p.Close()

	fmt.Println("所有任务已完成")
}

// processTask 模拟处理任务
func processTask(id int, priority string) error {
	fmt.Printf("开始处理任务 %d (优先级: %s)\n", id, priority)

	// 随机工作时间，模拟不同任务的处理时间
	workTime := time.Duration(rand.Intn(2)+1) * time.Second
	time.Sleep(workTime)

	fmt.Printf("任务 %d (优先级: %s) 已完成，耗时: %v\n", id, priority, workTime)
	return nil
}
