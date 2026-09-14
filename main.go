package main

import "fmt"

func twoSum(nums []int, target int) []int {

	visited := make(map[int]int, len(nums))

	for index, value := range nums {
		need := target - value

		countIndex, ok := visited[need]
		if ok {
			return []int{countIndex, index}
		}

		visited[value] = index
	}

	return []int{}
}

func main() {
	fmt.Println(twoSum([]int{1, 2, 3, 4, 1, 1, 1}, 7))
	fmt.Println(twoSum([]int{1, 2, 3, 4, 1, 1, 1}, 2))
	fmt.Println(twoSum([]int{7, 7, 7}, 9))
	fmt.Println(twoSum([]int{-1, -2, 56, 92, 0, 0}, -3))
	fmt.Println(twoSum([]int{3, 2, 4}, 6))
	fmt.Println(twoSum([]int{3, 3}, 6))
	fmt.Println(twoSum([]int{0, 4, 3, 0}, 0))
	fmt.Println(twoSum([]int{0, 7, 11, 0}, 7))
	fmt.Println(twoSum([]int{-10, 7, 19, 15}, 9))
	fmt.Println(twoSum([]int{-5, -2, -3, 1}, -8))
	fmt.Println(twoSum([]int{7, 7, 7}, 9))
	fmt.Println(twoSum([]int{1, 2, 3}, 7))
	fmt.Println(twoSum([]int{5}, 5))
	fmt.Println(twoSum([]int{}, 0))
	fmt.Println(twoSum([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 19))
}
