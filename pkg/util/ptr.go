package util

import "time"

// IntPtr returns a pointer to the input value.
func IntPtr(v int) *int {
	return &v
}

// StringPtr returns a pointer to the input value.
func StringPtr(v string) *string {
	return &v
}

// TimePtr returns a pointer to the input value.
func TimePtr(v time.Time) *time.Time {
	return &v
}
