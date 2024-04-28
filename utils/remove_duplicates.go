package utils

func RemoveDuplicates(values []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range values {
		if encountered[values[v]] {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[values[v]] = true
			// Append to result slice.
			result = append(result, values[v])
		}
	}
	return result
}
