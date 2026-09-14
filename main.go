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
}
