package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/shrimps80/go-service-utils/pool"
)

// 计算结果
type Result struct {
	TaskID int
	Value  int
	Time   time.Duration
}

func main() {
	// 创建一个大小为3的协程池
	p, err := pool.New(3)
	if err != nil {
		log.Fatalf("创建协程池失败: %v", err)
	}

	// 存储所有Future对象
	futures := make([]pool.Future, 0, 10)

	// 提交10个任务
	for i := 1; i <= 10; i++ {
		taskID := i
		future := p.SubmitFunc(func() (Result, error) {
			return calculateTask(taskID)
		})
		futures = append(futures, future)
	}

	// 创建上下文，用于获取结果
	ctx := context.Background()

	// 获取所有任务结果
	var totalValue int
	for i, future := range futures {
		result, err := future.Get(ctx)
		if err != nil {
			log.Printf("获取任务 %d 结果失败: %v", i+1, err)
			continue
		}

		// 类型断言获取结果
		if r, ok := result.(Result); ok {
			fmt.Printf("任务 %d 结果: 值=%d, 耗时=%v\n", r.TaskID, r.Value, r.Time)
			totalValue += r.Value
		}
	}

	// 获取统计信息
	stats := p.Stats()
	fmt.Printf("协程池统计: 运行任务=%d, 等待任务=%d, 已完成任务=%d\n",
		stats.RunningTasks, stats.WaitingTasks, stats.CompletedTasks)

	// 关闭协程池
	p.Close()

	fmt.Printf("所有任务已完成，结果总和: %d\n", totalValue)
}

// calculateTask 模拟计算任务
func calculateTask(id int) (Result, error) {
	fmt.Printf("开始计算任务 %d\n", id)

	// 随机工作时间，模拟不同任务的处理时间
	workTime := time.Duration(rand.Intn(3)+1) * time.Second
	time.Sleep(workTime)

	// 计算结果（简单示例：id的平方）
	value := id * id

	fmt.Printf("任务 %d 计算完成，结果: %d\n", id, value)

	return Result{
		TaskID: id,
		Value:  value,
		Time:   workTime,
	}, nil
}
