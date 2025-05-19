package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/shrimps80/go-service-utils/pool"
)

func main() {
	// 创建一个大小为4的协程池
	p, err := pool.New(4)
	if err != nil {
		log.Fatalf("创建协程池失败: %v", err)
	}

	// 随机种子
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 提交12个任务，设置不同的超时时间
	for i := 1; i <= 12; i++ {
		taskID := i

		// 随机工作时间，1-5秒
		workTime := time.Duration(rand.Intn(5)+1) * time.Second

		// 设置超时时间，2秒
		timeout := 2 * time.Second

		err := p.SubmitWithOptions(func() error {
			return processTask(taskID, workTime)
		}, pool.WithTimeout(timeout))

		if err != nil {
			log.Printf("提交任务 %d 失败: %v", taskID, err)
			continue
		}

		fmt.Printf("提交任务 %d (工作时间: %v, 超时时间: %v)\n", taskID, workTime, timeout)
	}

	// 等待所有任务完成
	p.Wait()

	// 获取统计信息
	stats := p.Stats()
	fmt.Printf("协程池统计: 运行任务=%d, 等待任务=%d, 已完成任务=%d, 超时任务=%d\n",
		stats.RunningTasks, stats.WaitingTasks, stats.CompletedTasks, stats.TimeoutTasks)

	// 关闭协程池
	p.Close()

	fmt.Println("所有任务已完成")
}

// processTask 模拟处理任务
func processTask(id int, workTime time.Duration) error {
	fmt.Printf("开始处理任务 %d (预计耗时: %v)\n", id, workTime)

	// 模拟工作
	time.Sleep(workTime)

	fmt.Printf("任务 %d 已完成，实际耗时: %v\n", id, workTime)
	return nil
}
