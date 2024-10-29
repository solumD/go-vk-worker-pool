package main

import (
	"fmt"

	"github.com/solumD/go-vk-worker-pool/pool"
)

func main() {
	strings := []string{"aaaaa", "bbbbb", "ccccc", "ddddd", "eeeee", "fffff"}

	p := pool.New()

	// добавляем воркеры
	p.Add()
	p.Add()

	fmt.Println("Пример 1")

	// запуск обработки, в выводе должны быть номера 1 и 2
	if err := p.StartProcessing(strings); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\nПример 2")

	// удаляем воркера
	if err := p.Remove(); err != nil {
		fmt.Println(err)
	}

	// еще раз запускаем обработку, в выводе должен быть только номер 1
	if err := p.StartProcessing(strings); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\nПример 3")

	// еще раз удаляем воркера (последнего)
	if err := p.Remove(); err != nil {
		fmt.Println(err)
	}

	// последний раз пытаемся запустить обработку, в выводе
	// должна быть ошибка о недостаточном кол-ве воркеров
	if err := p.StartProcessing(strings); err != nil {
		fmt.Println(err)
	}
}
