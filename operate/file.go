package operate

import (
	"io"
)

type FileBase interface {
	io.ReadWriter
	io.ReaderAt
	io.WriterAt
}
