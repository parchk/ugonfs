package comm

import (
	"errors"
)

var (
	ENOENT = errors.New("no such file or directory")
	ENOFH  = errors.New("fh not found")
)
