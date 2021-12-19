package goslice

func InSliceString(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}
func FilterSliceInt(sl []int) []int {
	var newSlice []int
	for _, s := range sl {
		if s != 0 {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}

func InSliceInt(v int, sl []int) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func SliceIntersectInt(slice1, slice2 []int) (diffslice []int) {
	for _, v := range slice1 {
		if InSliceInt(v, slice2) {
			diffslice = append(diffslice, v)
		}
	}
	return
}
