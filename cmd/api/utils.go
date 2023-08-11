package main

func Includes(input []int, key int) bool {
	var exists bool
	for _, v := range input {
		if v == key {
			exists = true
			break
		}
	}
	return exists
}
