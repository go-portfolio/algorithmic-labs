package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// 1. Интерфейс задачи - позволяет обрабатывать разные типы данных одним воркером
type Task interface {
	Process(ctx context.Context) (string, error)
	GetID() int
}

// --- Примеры различных задач ---

type HTTPTask struct{ ID int }
func (t HTTPTask) GetID() int { return t.ID }
func (t HTTPTask) Process(ctx context.Context) (string, error) {
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond) // Имитация сети
	return fmt.Sprintf("HTTP результат для %d", t.ID), nil
}

type DBTask struct{ ID int }
func (t DBTask) GetID() int { return t.ID }
func (t DBTask) Process(ctx context.Context) (string, error) {
	if t.ID%5 == 0 { return "", errors.New("ошибка БД") } // Имитация ошибки
	return fmt.Sprintf("DB данные для %d", t.ID), nil
}

// 2. Диспетчер системы
type Pool struct {
	tasks       chan Task
	results     chan string
	concurrency int
	wg          sync.WaitGroup
}

func NewPool(concurrency int) *Pool {
	return &Pool{
		tasks:       make(chan Task, 100),
		results:     make(chan string, 100),
		concurrency: concurrency,
	}
}

// 3. Логика воркера (Многопоточность + Интерфейсы)
func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok { return }
			
			// Выполнение через интерфейс
			res, err := task.Process(ctx)
			if err != nil {
				p.results <- fmt.Sprintf("[Ошибка] Задача %d: %v", task.GetID(), err)
			} else {
				p.results <- res
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool := NewPool(3) // Запускаем 3 параллельных потока

	// Запуск воркеров
	for i := 0; i < pool.concurrency; i++ {
		pool.wg.Add(1)
		go pool.worker(ctx)
	}

	// 4. Горутина-генератор задач (имитация входного потока)
	go func() {
		for i := 1; i <= 10; i++ {
			if i%2 == 0 {
				pool.tasks <- HTTPTask{ID: i}
			} else {
				pool.tasks <- DBTask{ID: i}
			}
		}
		close(pool.tasks)
	}()

	// 5. Горутина для закрытия канала результатов
	go func() {
		pool.wg.Wait()
		close(pool.results)
	}()

	// Сбор результатов (Pipeline stage)
	fmt.Println("Начинаем обработку...")
	for res := range pool.results {
		fmt.Println("<- Получено:", res)
	}
	fmt.Println("Все задачи завершены.")
}
