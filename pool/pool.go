// Package pool 提供了一个简单而强大的协程池实现，用于控制并发任务数量
package pool

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Task 表示要在协程池中执行的任务
type Task func() error

// Priority 任务优先级
type Priority int

// 预定义优先级常量
const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 5
	PriorityHigh   Priority = 10
)

// 错误定义
var (
	ErrPoolClosed      = errors.New("协程池已关闭")
	ErrContextCanceled = errors.New("任务上下文已取消")
	ErrTaskTimeout     = errors.New("任务执行超时")
	ErrNilTask         = errors.New("任务不能为空")
	ErrInvalidTask     = errors.New("无效的任务类型")
)

// Pool 协程池接口
type Pool interface {
	// Submit 提交一个任务到协程池
	Submit(task Task) error

	// SubmitWithContext 提交一个带上下文的任务到协程池
	SubmitWithContext(ctx context.Context, task Task) error

	// SubmitWithOptions 提交一个带选项的任务到协程池
	SubmitWithOptions(task Task, options ...TaskOption) error

	// SubmitFunc 提交一个带结果的任务到协程池，返回Future对象
	SubmitFunc(task interface{}) Future

	// SubmitFuncWithContext 提交一个带上下文和结果的任务到协程池，返回Future对象
	SubmitFuncWithContext(ctx context.Context, task interface{}) Future

	// Wait 等待所有任务完成
	Wait()

	// Close 关闭协程池，不再接受新任务
	Close()

	// IsClosed 检查协程池是否已关闭
	IsClosed() bool

	// Stats 返回协程池的统计信息
	Stats() Stats
}

// Stats 协程池统计信息
type Stats struct {
	// Size 协程池大小（最大并发数）
	Size int

	// RunningTasks 当前正在运行的任务数
	RunningTasks int

	// WaitingTasks 当前等待中的任务数
	WaitingTasks int

	// CompletedTasks 已完成的任务数
	CompletedTasks int64

	// TimeoutTasks 超时的任务数
	TimeoutTasks int64
}

// Future 表示一个异步任务的未来结果
type Future interface {
	// Get 获取任务结果，阻塞直到任务完成或上下文取消
	Get(ctx context.Context) (interface{}, error)

	// GetWithTimeout 获取任务结果，阻塞直到任务完成、超时或上下文取消
	GetWithTimeout(timeout time.Duration) (interface{}, error)

	// IsDone 检查任务是否已完成
	IsDone() bool
}

// Options 协程池选项
type Options struct {
	// QueueSize 任务队列大小，默认为 size * 10
	QueueSize int

	// EnablePriority 是否启用优先级功能
	EnablePriority bool

	// PanicHandler 处理任务中的 panic
	PanicHandler func(interface{})
}

// Option 协程池选项函数
type Option func(*Options)

// WithQueueSize 设置任务队列大小
func WithQueueSize(size int) Option {
	return func(o *Options) {
		o.QueueSize = size
	}
}

// WithPriority 启用优先级功能
func WithPriority() Option {
	return func(o *Options) {
		o.EnablePriority = true
	}
}

// WithPanicHandler 设置 panic 处理函数
func WithPanicHandler(handler func(interface{})) Option {
	return func(o *Options) {
		o.PanicHandler = handler
	}
}

// TaskOptions 任务选项
type TaskOptions struct {
	// Priority 任务优先级，默认为 PriorityNormal
	Priority Priority

	// Timeout 任务超时时间，默认为 0（不超时）
	Timeout time.Duration
}

// TaskOption 任务选项函数
type TaskOption func(*TaskOptions)

// WithTaskPriority 设置任务优先级
func WithTaskPriority(priority Priority) TaskOption {
	return func(o *TaskOptions) {
		o.Priority = priority
	}
}

// WithTimeout 设置任务超时时间
func WithTimeout(timeout time.Duration) TaskOption {
	return func(o *TaskOptions) {
		o.Timeout = timeout
	}
}

// poolImpl 协程池实现
type poolImpl struct {
	size           int              // 协程池大小（最大并发数）
	taskQueue      chan taskWrapper // 普通任务队列
	priorityQueue  *priorityQueue   // 优先级队列
	priorityLock   sync.Mutex       // 优先级队列锁
	wg             sync.WaitGroup   // 用于等待所有任务完成
	closed         int32            // 协程池是否已关闭
	runningTasks   int32            // 当前正在运行的任务数
	completedTasks int64            // 已完成的任务数
	timeoutTasks   int64            // 超时的任务数
	dispatcherWg   sync.WaitGroup   // 用于等待调度器协程完成
	options        Options          // 协程池选项
}

// taskWrapper 任务包装器
type taskWrapper struct {
	task     interface{}
	ctx      context.Context
	priority Priority
	timeout  time.Duration
	added    time.Time
	future   *futureImpl
	index    int // 在堆中的索引
}

// futureImpl Future接口实现
type futureImpl struct {
	result     interface{}
	err        error
	done       bool
	mu         sync.Mutex
	completeCh chan struct{}
}

// priorityQueue 优先级队列
type priorityQueue []*taskWrapper

// New 创建一个新的协程池
func New(size int, opts ...Option) (Pool, error) {
	if size <= 0 {
		return nil, fmt.Errorf("协程池大小必须大于0，当前值: %d", size)
	}

	// 默认选项
	options := Options{
		QueueSize:      size * 10,
		EnablePriority: false,
		PanicHandler: func(p interface{}) {
			fmt.Printf("协程池任务发生panic: %v\n", p)
		},
	}

	// 应用选项
	for _, opt := range opts {
		opt(&options)
	}

	p := &poolImpl{
		size:    size,
		options: options,
	}

	// 根据是否启用优先级功能，初始化不同的任务队列
	if options.EnablePriority {
		pq := make(priorityQueue, 0)
		heap.Init(&pq)
		p.priorityQueue = &pq

		// 启动调度器协程
		p.dispatcherWg.Add(1)
		go p.dispatcher()
	} else {
		p.taskQueue = make(chan taskWrapper, options.QueueSize)
	}

	// 启动工作协程
	for i := 0; i < size; i++ {
		go p.worker()
	}

	return p, nil
}

// Submit 提交一个任务到协程池
func (p *poolImpl) Submit(task Task) error {
	return p.SubmitWithContext(context.Background(), task)
}

// SubmitWithContext 提交一个带上下文的任务到协程池
func (p *poolImpl) SubmitWithContext(ctx context.Context, task Task) error {
	return p.SubmitWithOptions(task, func(o *TaskOptions) {
		// 使用默认选项
		o.Priority = PriorityNormal
		o.Timeout = 0
	})
}

// SubmitWithOptions 提交一个带选项的任务到协程池
func (p *poolImpl) SubmitWithOptions(task Task, options ...TaskOption) error {
	if task == nil {
		return ErrNilTask
	}

	if p.IsClosed() {
		return ErrPoolClosed
	}

	// 默认任务选项
	taskOpts := TaskOptions{
		Priority: PriorityNormal,
		Timeout:  0,
	}

	// 应用任务选项
	for _, opt := range options {
		opt(&taskOpts)
	}

	// 创建任务包装器
	tw := taskWrapper{
		task:     task,
		ctx:      context.Background(),
		priority: taskOpts.Priority,
		timeout:  taskOpts.Timeout,
		added:    time.Now(),
	}

	return p.submitTask(tw)
}

// SubmitFunc 提交一个带结果的任务到协程池，返回Future对象
func (p *poolImpl) SubmitFunc(task interface{}) Future {
	return p.SubmitFuncWithContext(context.Background(), task)
}

// SubmitFuncWithContext 提交一个带上下文和结果的任务到协程池，返回Future对象
func (p *poolImpl) SubmitFuncWithContext(ctx context.Context, task interface{}) Future {
	// 检查任务类型
	if task == nil {
		return newErrorFuture(ErrNilTask)
	}

	if p.IsClosed() {
		return newErrorFuture(ErrPoolClosed)
	}

	// 创建Future对象
	future := &futureImpl{
		completeCh: make(chan struct{}),
	}

	// 创建任务包装器
	tw := taskWrapper{
		task:     task,
		ctx:      ctx,
		priority: PriorityNormal,
		timeout:  0,
		added:    time.Now(),
		future:   future,
	}

	// 提交任务
	err := p.submitTask(tw)
	if err != nil {
		return newErrorFuture(err)
	}

	return future
}

// submitTask 提交任务到队列
func (p *poolImpl) submitTask(tw taskWrapper) error {
	// 检查上下文是否已取消
	select {
	case <-tw.ctx.Done():
		return ErrContextCanceled
	default:
	}

	p.wg.Add(1)

	// 根据是否启用优先级功能，使用不同的任务队列
	if p.options.EnablePriority {
		// 将任务添加到优先级队列
		p.priorityLock.Lock()
		heap.Push(p.priorityQueue, &tw)
		p.priorityLock.Unlock()
		return nil
	} else {
		// 将任务添加到普通队列
		select {
		case p.taskQueue <- tw:
			return nil
		case <-tw.ctx.Done():
			p.wg.Done()
			return ErrContextCanceled
		}
	}
}

// dispatcher 任务调度器（仅在启用优先级功能时使用）
func (p *poolImpl) dispatcher() {
	defer p.dispatcherWg.Done()

	// 创建工作通道
	workCh := make(chan taskWrapper, p.size)
	defer close(workCh)

	// 启动工作协程
	for i := 0; i < p.size; i++ {
		go p.priorityWorker(workCh)
	}

	for !p.IsClosed() || p.priorityQueueLen() > 0 {
		p.priorityLock.Lock()
		if p.priorityQueue.Len() == 0 {
			p.priorityLock.Unlock()
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// 获取最高优先级的任务
		tw := heap.Pop(p.priorityQueue).(*taskWrapper)
		p.priorityLock.Unlock()

		// 检查上下文是否已取消
		select {
		case <-tw.ctx.Done():
			if tw.future != nil {
				tw.future.setError(ErrContextCanceled)
			}
			p.wg.Done() // 任务被取消，直接标记完成
			continue
		default:
		}

		// 将任务发送到工作通道
		select {
		case workCh <- *tw:
			// 任务已发送
		case <-tw.ctx.Done():
			if tw.future != nil {
				tw.future.setError(ErrContextCanceled)
			}
			p.wg.Done() // 任务被取消，直接标记完成
		}
	}
}

// worker 普通工作协程
func (p *poolImpl) worker() {
	for tw := range p.taskQueue {
		p.processTask(tw)
	}
}

// priorityWorker 优先级工作协程
func (p *poolImpl) priorityWorker(workCh <-chan taskWrapper) {
	for tw := range workCh {
		p.processTask(tw)
	}
}

// processTask 处理任务
func (p *poolImpl) processTask(tw taskWrapper) {
	// 检查上下文是否已取消
	select {
	case <-tw.ctx.Done():
		if tw.future != nil {
			tw.future.setError(ErrContextCanceled)
		}
		p.wg.Done()
		return
	default:
	}

	// 增加运行中的任务计数
	atomic.AddInt32(&p.runningTasks, 1)
	defer atomic.AddInt32(&p.runningTasks, -1)

	// 创建带超时的上下文（如果需要）
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if tw.timeout > 0 {
		ctx, cancel = context.WithTimeout(tw.ctx, tw.timeout)
	} else {
		ctx, cancel = context.WithCancel(tw.ctx)
	}
	defer cancel()

	// 执行任务
	done := make(chan struct{})
	var result interface{}
	var err error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("任务panic: %v", r)
				if p.options.PanicHandler != nil {
					p.options.PanicHandler(r)
				}
			}
			close(done)
		}()

		// 根据任务类型执行不同的处理
		if task, ok := tw.task.(Task); ok {
			err = task()
		} else {
			// 使用反射调用函数并获取结果
			result, err = p.callFunc(tw.task)
		}
	}()

	// 等待任务完成或超时
	select {
	case <-done:
		// 任务正常完成
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			// 任务超时
			err = ErrTaskTimeout
			atomic.AddInt64(&p.timeoutTasks, 1)
		} else {
			// 上下文取消
			err = ErrContextCanceled
		}
	}

	// 设置Future结果（如果有）
	if tw.future != nil {
		if err != nil {
			tw.future.setError(err)
		} else {
			tw.future.setResult(result, nil)
		}
	}

	// 增加已完成任务计数
	atomic.AddInt64(&p.completedTasks, 1)

	// 标记任务完成
	p.wg.Done()
}

// callFunc 使用反射调用函数并获取结果
func (p *poolImpl) callFunc(fn interface{}) (interface{}, error) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return nil, ErrInvalidTask
	}

	t := v.Type()
	if t.NumIn() != 0 {
		return nil, fmt.Errorf("任务函数不应有参数")
	}

	if t.NumOut() == 0 {
		v.Call(nil)
		return nil, nil
	}

	results := v.Call(nil)
	if len(results) == 1 {
		return results[0].Interface(), nil
	}

	// 处理返回 (T, error) 的情况
	if len(results) == 2 && t.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !results[1].IsNil() {
			return results[0].Interface(), results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	}

	return results, nil
}

// Wait 等待所有任务完成
func (p *poolImpl) Wait() {
	p.wg.Wait()
}

// Close 关闭协程池，不再接受新任务
func (p *poolImpl) Close() {
	// 原子操作设置关闭标志
	if atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		if p.options.EnablePriority {
			// 等待调度器完成
			p.dispatcherWg.Wait()
		} else {
			// 关闭任务队列
			close(p.taskQueue)
		}
	}
}

// IsClosed 检查协程池是否已关闭
func (p *poolImpl) IsClosed() bool {
	return atomic.LoadInt32(&p.closed) == 1
}

// Stats 返回协程池的统计信息
func (p *poolImpl) Stats() Stats {
	stats := Stats{
		Size:           p.size,
		RunningTasks:   int(atomic.LoadInt32(&p.runningTasks)),
		CompletedTasks: atomic.LoadInt64(&p.completedTasks),
		TimeoutTasks:   atomic.LoadInt64(&p.timeoutTasks),
	}

	if p.options.EnablePriority {
		stats.WaitingTasks = p.priorityQueueLen()
	} else {
		stats.WaitingTasks = len(p.taskQueue)
	}

	return stats
}

// priorityQueueLen 获取优先级队列长度
func (p *poolImpl) priorityQueueLen() int {
	p.priorityLock.Lock()
	defer p.priorityLock.Unlock()
	return p.priorityQueue.Len()
}

// newErrorFuture 创建一个带错误的Future
func newErrorFuture(err error) Future {
	f := &futureImpl{
		err:        err,
		done:       true,
		completeCh: make(chan struct{}),
	}
	close(f.completeCh)
	return f
}

// setResult 设置Future结果
func (f *futureImpl) setResult(result interface{}, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.done {
		f.result = result
		f.err = err
		f.done = true
		close(f.completeCh)
	}
}

// setError 设置Future错误
func (f *futureImpl) setError(err error) {
	f.setResult(nil, err)
}

// Get 获取Future结果
func (f *futureImpl) Get(ctx context.Context) (interface{}, error) {
	// 检查任务是否已完成
	f.mu.Lock()
	if f.done {
		result := f.result
		err := f.err
		f.mu.Unlock()
		return result, err
	}
	f.mu.Unlock()

	// 等待任务完成或上下文取消
	select {
	case <-f.completeCh:
		f.mu.Lock()
		result := f.result
		err := f.err
		f.mu.Unlock()
		return result, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetWithTimeout 获取Future结果，带超时
func (f *futureImpl) GetWithTimeout(timeout time.Duration) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return f.Get(ctx)
}

// IsDone 检查Future是否已完成
func (f *futureImpl) IsDone() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.done
}

// 以下是优先级队列所需的接口实现

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// 优先级高的排在前面
	if pq[i].priority != pq[j].priority {
		return pq[i].priority > pq[j].priority
	}
	// 同优先级按添加时间排序（先进先出）
	return pq[i].added.Before(pq[j].added)
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*taskWrapper)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // 避免内存泄漏
	item.index = -1 // 标记为已移除
	*pq = old[0 : n-1]
	return item
}
