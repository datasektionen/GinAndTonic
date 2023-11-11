package utils

import "strconv"

func ParseStringToUint(s string) (uint, error) {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
