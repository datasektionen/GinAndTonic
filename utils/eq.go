package utils

import "time"

func IsEqualTimePtr(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b // true if both are nil, false if only one is nil
	}
	return a.Equal(*b)
}
