package parser

import (
	"fmt"
	"io"
)

type FlatExtentHeader struct {
	Reader  io.ReaderAt
	Offset  int64
	Profile *VMDKProfile
}

type FlatExtent struct {
	profile *VMDKProfile
	reader  io.ReaderAt

	header *FlatExtentHeader

	total_size int64

	// The offset in the logical image where this extent sits.
	offset   int64
	filename string

	closer func()
}

func (self *FlatExtent) Close() {
	if self.closer != nil {
		self.closer()
	}
}

func (self *FlatExtent) TotalSize() int64 {
	return self.total_size
}

func (self *FlatExtent) VirtualOffset() int64 {
	return self.offset
}

func (self *FlatExtent) ReadAt(buf []byte, offset int64) (int, error) {
	if offset < 0 || offset >= self.total_size {
		return 0, io.EOF
	}

	toRead := int64(len(buf))
	if offset+toRead > self.total_size {
		toRead = self.total_size - offset
	}

	fileOffset := self.header.Offset + offset
	return self.reader.ReadAt(buf[:toRead], fileOffset)
}

func (self *FlatExtent) Stats() ExtentStat {
	return ExtentStat{
		Type:     "flat",
		Size:     self.total_size,
		Filename: self.filename,
	}
}

func (self *FlatExtent) Debug() {
	fmt.Printf("[FlatExtent] file: %s, offset: %d, size: %d\n", self.filename, self.offset, self.total_size)
}

func GetFlatExtent(
	reader io.ReaderAt,
	filename string,
	offsetSectors int64,
	sectors int64,
	virtualOffset int64,
	profile *VMDKProfile,
	closer func(),
) (Extent, error) {
	flatExtentHeader := &FlatExtentHeader{
		Reader:  reader,
		Offset:  offsetSectors * SECTOR_SIZE,
		Profile: profile,
	}

	res := &FlatExtent{
		profile:    profile,
		reader:     reader,
		header:     flatExtentHeader,
		offset:     virtualOffset,
		total_size: sectors * SECTOR_SIZE,
		filename:   filename,
		closer:     closer,
	}
	return res, nil
}
