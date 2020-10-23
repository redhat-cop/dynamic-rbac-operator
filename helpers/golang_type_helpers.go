package helpers

func stringInSlice(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func slicesIntersect(s1, s2 []string) bool {
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[e] = true
	}
	for _, e := range s2 {
		if hash[e] {
			return true
		}
	}
	return false
}

func subtractStringSlices(allElements []string, elementsToRemove []string) []string {
	newList := []string{}
	for _, element := range allElements {
		if !stringInSlice(elementsToRemove, element) {
			newList = append(newList, element)
		}
	}
	return newList
}
