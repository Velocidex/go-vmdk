package parser

import "io"

type NullExtent struct {
	SparseExtent
}

func (self *NullExtent) ReadAt(buf []byte, offset int64) (int, error) {
	if offset < 0 || offset > self.total_size {
		return 0, io.EOF
	}

	to_read := int64(len(buf))
	available_length := self.total_size - offset
	if to_read > available_length {
		to_read = available_length
	}

	for i := int64(0); i < to_read; i++ {
		buf[i] = 0
	}

	return int(to_read), nil
}

func (self *NullExtent) Stats() ExtentStat {
	return ExtentStat{
		Type:          "PAD",
		VirtualOffset: self.offset,
		Size:          self.total_size,
		Filename:      self.filename,
	}
}
