package parse

import (
	"fmt"
	"strconv"
)

func PositiveInt(num string) (int, error) {
	n, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, fmt.Errorf("must be a positive integer")
	}
	return n, nil
}
