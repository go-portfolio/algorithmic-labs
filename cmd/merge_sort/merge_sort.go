package main

import (
	"cmp"
	"fmt"
)

// MergeSort — основная функция, которая рекурсивно делит слайс
func MergeSort[T cmp.Ordered](items []T) []T {
	if len(items) <= 1 {
		return items
	}

	middle := len(items) / 2
	
	// Рекурсивно вызываем для левой и правой половин
	left := MergeSort(items[:middle])
	right := MergeSort(items[middle:])

	// Объединяем результаты
	return merge(left, right)
}

// merge — вспомогательная функция для слияния двух отсортированных слайсов
func merge[T cmp.Ordered](left, right []T) []T {
	result := make([]T, 0, len(left)+len(right))
	i, j := 0, 0

	// Сравниваем элементы из обеих половин и добавляем меньший в результат
	for i < len(left) && j < len(right) {
		if left[i] < right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}

	// Добавляем оставшиеся элементы (если они есть)
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)

	return result
}

func main() {
	unsorted := []int{38, 27, 43, 3, 9, 82, 10}
	sorted := MergeSort(unsorted)
	fmt.Println("Отсортированный массив:", sorted)
}
