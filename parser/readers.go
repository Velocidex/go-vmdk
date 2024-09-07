package parser

import "io"

type Extent interface {
	io.ReaderAt

	VirtualOffset() int64
	TotalSize() int64
	Stats() ExtentStat
	Close()
	Debug()
}
