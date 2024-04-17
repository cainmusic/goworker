package goworker

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

const MaxWorkerNumber = 10000

var (
	WorkerNumberTooSmall = errors.New("worker number must > 0")
	WorkerNumberTooLarge = errors.New(fmt.Sprintf("worker number must <= %d", MaxWorkerNumber))
	TaskQueueDone        = errors.New("task queue done, no more task")
)

type task func()

type flag struct {
	fg bool
	mu sync.Mutex
}

func (f *flag) set(fg bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fg = fg
}

func (f *flag) get() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.fg
}

type manager struct {
	wg *sync.WaitGroup // wait group
	ws []*worker       // workers
	tq chan task       // task queue
	df *flag           // done flag
}

func newManager(workerNumber int) (*manager, error) {
	err := checkWorkerNumber(workerNumber)
	if err != nil {
		return nil, err
	}

	m := &manager{
		// 传给worker使用，使用传递指针，用来确保worker已退出
		wg: &sync.WaitGroup{},
		// 需要传给worker使用，worker会从中获取任务，缓冲区设为worker数量的10倍
		tq: make(chan task, workerNumber*10),
		// 用来标记停止向队列中发送任务
		df: &flag{},
	}

	m.wg.Add(workerNumber)

	workers := make([]*worker, workerNumber)
	for i, _ := range workers {
		wid := i
		workers[i] = newWorker(m.wg, m.tq, wid)
	}

	m.ws = workers

	return m, nil
}

func checkWorkerNumber(workerNumber int) error {
	if workerNumber <= 0 {
		return WorkerNumberTooSmall
	}
	if workerNumber > MaxWorkerNumber {
		return WorkerNumberTooLarge
	}
	return nil
}

func (m *manager) AddTask(f task) error {
	if m.df.get() {
		return TaskQueueDone
	}
	m.tq <- f
	return nil
}

func (m *manager) Done() {
	// 停止接收任务
	m.df.set(true)

	// 关闭task queue即相当于通知worker结束
	close(m.tq)

	// 等待workers结束
	log.Println("waiting for all workers done")
	m.wg.Wait()
}

func NewPool(workerNumber int) (*manager, error) {
	return newManager(workerNumber)
}

type worker struct {
	wg *sync.WaitGroup // manager.wg
	tq chan task       // manager.tq
	id int             // id
}

func newWorker(wg *sync.WaitGroup, tq chan task, id int) *worker {
	w := &worker{
		wg: wg,
		tq: tq,
		id: id,
	}

	go w.run()

	return w
}

func (w *worker) run() {
	// 使用range可以在w.tq被close之后，将w.tq排空后自动退出
	for f := range w.tq {
		log.Println("doing task")
		f()
	}
	log.Printf("worker %d done", w.id)
	w.wg.Done()
}
