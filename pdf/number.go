package pdf

import (
	"fmt"
)

type Number float64

func (number Number) String() string {
	return fmt.Sprint(float64(number))
}
