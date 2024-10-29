package pool

import (
	"fmt"
	"sync"
)

// Worker печатает в консоль строку и номер воркера
func Worker(str string, workerNumber int) {
	fmt.Printf("Обрабатываемая строка: %s, номер воркера: %d\n", str, workerNumber)
}

// WorkerkPool пул из workersCount воркеров
type WorkerPool struct {
	mu           *sync.Mutex
	workersCount int
}

// New возвращает новый объект WorkerPool
func New() *WorkerPool {
	return &WorkerPool{
		mu:           &sync.Mutex{},
		workersCount: 0,
	}
}

// Add добавляет воркер в пул (увеличивает их счетчик на 1)
func (p *WorkerPool) Add() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.workersCount++
}

// Remove удаляет воркер из пула (уменьшает их счетчик на 1),
// если их количество больше нуля, иначе возвращает ошибку
func (p *WorkerPool) Remove() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.workersCount == 0 {
		return fmt.Errorf("удаление воркера не удалось: текущее кол-во воркеров равно 0")
	}

	p.workersCount--
	return nil
}

// StarProcessing запускает обработку полученного слайса строк, если кол-во воркеров больше нуля,
// иначе возвращает ошибку
func (p *WorkerPool) StartProcessing(strs []string) error {
	p.mu.Lock()
	if p.workersCount == 0 {
		return fmt.Errorf("запуск обработки не удался: кол-во воркеров равно 0")
	}
	p.mu.Unlock()

	// создаем пул воркеров и заполняем его "токенами" (i)
	pool := make(chan int, p.workersCount)
	for i := 1; i <= p.workersCount; i++ {
		pool <- i
	}

	// канал, куда будем записывать полученные строки
	strChan := make(chan string)

	// в горутине, пишем в канал строки из слайса
	go func() {
		for _, s := range strs {
			strChan <- s
		}

		close(strChan)
	}()

	// читаем строки из канала
	for str := range strChan {
		workerID := <-pool // берем "токен"
		go func() {
			Worker(str, workerID)
			pool <- workerID // возвращаем токен в пул
		}()
	}

	// ждем, когда все токены вернуться в пул
	for i := 0; i < p.workersCount; i++ {
		<-pool
	}

	return nil
}
