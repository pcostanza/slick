package list

import (
	"fmt"
)

func outOfBounds(index int, list interface{}) error {
	return fmt.Errorf("index %v out of bounds for list %v", index, list)
}

func negativeLength(length int) error {
	return fmt.Errorf("negative length %v is invalid for lists", length)
}
