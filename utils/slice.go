package utils

func SliceDiff(a, b []int) []int {
	mb := make(map[int]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []int
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func MapDiff(a, b map[interface{}]interface{}) map[interface{}]interface{} {
	diff := make(map[interface{}]interface{})
	for x := range a {
		if val, ok := b[x]; !ok {
			diff[x] = val
		}
	}

	return diff
}

func Keys(a map[int]interface{}) []int {
	var keys []int
	for k := range a {
		keys = append(keys, k)
	}
	return keys
}

func Contains(array interface{}, item interface{}) (bool, int) {
	switch v := array.(type) {
	case []int64:
		for index, val := range v {
			if val == item {
				return true, index
			}
		}
	}

	return false, -1
}

func CreateBytes(value byte, length, size int) []byte {
	arr := make([]byte, size)
	for i := 0; i < length; i++ {
		arr[i] = value
	}

	return arr
}

func CreateInts(value, length, size int) []int {
	arr := make([]int, size)
	for i := 0; i < length; i++ {
		arr[i] = value
	}

	return arr
}

func SearchUInt64(a []uint64, x uint64) int {
	return search(len(a), func(i uint64) bool { return a[i] > x })
}

func search(n int, f func(uint64) bool) int {
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	i, j := uint64(0), uint64(n)
	for i < j {
		h := (i + j) >> 1
		// i ≤ h < j
		if !f(h) {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return int(i)
}
