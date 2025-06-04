package pool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var errPoolIsStopped = fmt.Errorf("pool is stopped")

// WorkerPool пул из workersCount воркеров
type WorkerPool struct {
	workersCount atomic.Int64
	stopped      bool
	tasks        chan string
	addSignal    chan struct{}
	removeSignal chan struct{}
	stopSignal   chan struct{}
	mu           *sync.Mutex
	wg           *sync.WaitGroup
}

// New возвращает новый объект WorkerPool с заданным количеством воркеров (по умолчанию 1) и минимальным буфером задач (по умолчанию 1)
func New(workersCount int, tasksBuffer int) *WorkerPool {
	if workersCount < 1 {
		workersCount = 1
	}

	if tasksBuffer < 1 {
		tasksBuffer = 1
	}

	pool := &WorkerPool{
		workersCount: atomic.Int64{},
		stopped:      false,
		tasks:        make(chan string, tasksBuffer),
		addSignal:    make(chan struct{}),
		removeSignal: make(chan struct{}),
		stopSignal:   make(chan struct{}),
		mu:           &sync.Mutex{},
		wg:           &sync.WaitGroup{},
	}

	pool.workersCount.Store(int64(workersCount))

	go pool.startProcessing()

	return pool
}

// AddWorker добавляет воркер в пул
func (wp *WorkerPool) AddWorker() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.stopped {
		return fmt.Errorf("failed to add worker: %s", errPoolIsStopped)
	}

	wp.workersCount.Add(1)
	wp.addSignal <- struct{}{}

	return nil
}

// RemoveWorker удаляет воркер из пула
func (wp *WorkerPool) RemoveWorker() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.stopped {
		return fmt.Errorf("failed to remove worker: %s", errPoolIsStopped)
	}

	// должен быть минимум один воркер для корректной работы
	if wp.workersCount.Load() == 1 {
		return errors.New("failed to remove worker: minimum workers count is 1")
	}

	wp.workersCount.Add(-1)
	wp.removeSignal <- struct{}{}

	return nil
}

// StopPool останавливает весь пул
func (wp *WorkerPool) StopPool() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.stopped {
		return fmt.Errorf("failed to stop pool: %s", errPoolIsStopped)
	}

	wp.stopped = true
	close(wp.stopSignal)
	close(wp.tasks)

	return nil
}

// SendTask отправляет задачу в пул
func (wp *WorkerPool) SendTask(task string) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.stopped {
		return fmt.Errorf("failed to send task: %s", errPoolIsStopped)
	}

	wp.tasks <- task
	return nil
}

// startWorker запускает новый воркер в пуле
func (wp *WorkerPool) startWorker(uuid string) {
	wp.wg.Add(1)
	go func() {
		defer wp.wg.Done()
		for {
			select {
			case <-wp.removeSignal:
				fmt.Printf("Worker %s was removed\n", uuid)
				return

			case <-wp.stopSignal:
				fmt.Printf("Worker %s was stopped\n", uuid)
				return

			case task, ok := <-wp.tasks:
				if !ok {
					fmt.Printf("Worker %s stopped: tasks channel closed\n", uuid)
					return
				}

				fmt.Printf("Tasks processed: %s. Worker's uuid: %s\n", task, uuid)
				// имитация обработки
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// startProcessing запускает пул
func (wp *WorkerPool) startProcessing() {
	// первый запуск
	for i := 0; i < int(wp.workersCount.Load()); i++ {
		wp.startWorker(uuid.New().String())
	}

	// далее слушаем каналы и в зависимости от сигнала добавляем или удаляем воркера
AddOrStopLoop:
	for {
		select {
		case <-wp.addSignal:
			wp.startWorker(uuid.New().String())

		case <-wp.stopSignal:
			break AddOrStopLoop
		}
	}

	wp.wg.Wait()
}
