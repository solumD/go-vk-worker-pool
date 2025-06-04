package main

import (
	"fmt"
	"time"

	"github.com/solumD/go-vk-worker-pool/pool"
)

func main() {
	// Создаем пул с 3 воркерами и буфером на 10 задач
	pool := pool.New(1, 1)

	// Запускаем горутину для отправки задач
	go func() {
		for i := 0; i < 20; i++ {
			task := fmt.Sprintf("task-%d", i)
			if err := pool.SendTask(task); err != nil {
				fmt.Printf("Error sending task %d: %v\n", i, err)
			}

			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Тестируем добавление/удаление воркеров
	time.Sleep(300 * time.Millisecond)
	pool.AddWorker()
	fmt.Println("Added worker")

	time.Sleep(300 * time.Millisecond)
	pool.RemoveWorker()
	fmt.Println("Removed worker")

	time.Sleep(300 * time.Millisecond)
	pool.AddWorker()
	fmt.Println("Added worker")

	// Даем время на обработку оставшихся задач
	time.Sleep(1 * time.Second)

	// Останавливаем пул
	pool.StopPool()
	fmt.Println("Pool stopped")

	// Пытаемся отправить задачу после остановки
	err := pool.SendTask("should-fail")
	if err != nil {
		fmt.Println(err)
	}
}
