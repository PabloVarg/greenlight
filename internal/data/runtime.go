package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("Invalid runtime format")

type Runtime int

func (r Runtime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(fmt.Sprintf("%d mins", r))), nil
}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	value, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	fields := strings.Fields(value)
	if len(fields) != 2 {
		return ErrInvalidRuntimeFormat
	}
	if fields[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	runtime, err := strconv.ParseInt(fields[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(runtime)

	return nil
}
