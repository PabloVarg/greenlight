package data

import (
	"fmt"
	"strconv"
)

type Runtime int

func (r Runtime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(fmt.Sprintf("%d mins", r))), nil
}
