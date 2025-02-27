package code_fun

func TopK(nums []int, k int) []int {
	if k < 1 || k > len(nums) {
		return nil
	}

	left, right := 0, len(nums)-1
	for {
		pivotIndex := partition(nums, left, right, right)
		if pivotIndex == k-1 {
			break
		} else if pivotIndex > k-1 {
			right = pivotIndex - 1
		} else {
			left = pivotIndex + 1
		}
	}

	return nums[:k]
}

func partition(nums []int, left, right, pivotIndex int) int {
	pivotValue := nums[pivotIndex]
	nums[pivotIndex], nums[right] = nums[right], nums[pivotIndex]
	idx := left

	for i := left; i < right; i++ {
		if nums[i] > pivotValue {
			nums[idx], nums[i] = nums[i], nums[idx]
			idx++
		}
	}

	nums[idx], nums[right] = nums[right], nums[idx]
	return idx
}
