package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// --- Код алгоритма (кратко) ---

type Task func()

type AdaptivePool struct {
	tasks       chan Task
	maxWorkers  int
	minWorkers  int
	currWorkers int
	mu          sync.Mutex
	stop        chan struct{}
}

func NewAdaptivePool(min, max int) *AdaptivePool {
	p := &AdaptivePool{
		tasks:      make(chan Task, 10), // Небольшой буфер для наглядности
		minWorkers: min,
		maxWorkers: max,
		stop:       make(chan struct{}),
	}
	for i := 0; i < min; i++ {
		p.addWorker()
	}
	go p.monitor()
	return p
}

func (p *AdaptivePool) addWorker() {
	p.mu.Lock()
	if p.currWorkers >= p.maxWorkers {
		p.mu.Unlock()
		return
	}
	p.currWorkers++
	curr := p.currWorkers
	p.mu.Unlock()

	fmt.Printf("[Pool] Запущен новый воркер. Всего: %d\n", curr)

	go func() {
		for {
			select {
			case task := <-p.tasks:
				task()
			case <-time.After(3 * time.Second): // Воркер умрет, если нет задач 3 сек
				p.mu.Lock()
				if p.currWorkers > p.minWorkers {
					p.currWorkers--
					fmt.Printf("[Pool] Воркер завершил работу по таймауту. Осталось: %d\n", p.currWorkers)
					p.mu.Unlock()
					return
				}
				p.mu.Unlock()
			case <-p.stop:
				return
			}
		}
	}()
}

func (p *AdaptivePool) monitor() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for range ticker.C {
		if len(p.tasks) > 5 { // Если в очереди больше 5 задач, расширяем пул
			p.addWorker()
		}
	}
}

func (p *AdaptivePool) Submit(t Task) error {
	select {
	case p.tasks <- t:
		return nil
	default:
		return fmt.Errorf("перегрузка! задача отклонена")
	}
}

// --- Точка входа ---

func main() {
	// Создаем пул: минимум 2 воркера, максимум 10
	pool := NewAdaptivePool(2, 10)

	// Имитируем поток из 30 входящих задач
	for i := 1; i <= 30; i++ {
		id := i
		task := func() {
			// Каждая задача выполняется от 500мс до 1.5с
			duration := time.Duration(rand.Intn(1000)+500) * time.Millisecond
			time.Sleep(duration)
			fmt.Printf("✅ Задача %d выполнена за %v\n", id, duration)
		}

		err := pool.Submit(task)
		if err != nil {
			fmt.Printf("❌ Задача %d: %v\n", id, err)
		}

		// Пауза между отправкой задач, чтобы не забить всё мгновенно
		time.Sleep(200 * time.Millisecond)
	}

	// Даем время доработать оставшимся задачам и воркерам — уйти по таймауту
	fmt.Println("--- Все задачи отправлены, ждем завершения ---")
	time.Sleep(15 * time.Second)
}
