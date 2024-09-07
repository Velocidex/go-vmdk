package parser

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

const (
	SPARSE_MAGICNUMBER = 0x564d444b
	SECTOR_SIZE        = 512
)

var (
	StartExtentRegex = regexp.MustCompile("^# Extent description")
	ExtentRegex      = regexp.MustCompile(`(RW|R) (\d+) ([A-Z]+) "([^"]+)"`)
)

type VMDKContext struct {
	profile *VMDKProfile
	reader  io.ReaderAt

	extents []*SparseExtent

	total_size int64
}

func (self *VMDKContext) Size() int64 {
	return self.total_size
}

func (self *VMDKContext) Debug() {
	for _, i := range self.extents {
		i.Debug()
	}
}

func (self *VMDKContext) Close() {
	for _, i := range self.extents {
		i.Close()
	}
}

func (self *VMDKContext) getGrainForOffset(offset int64) (
	reader io.ReaderAt, start, length int64, err error) {

	n, _ := slices.BinarySearchFunc(self.extents,
		offset, func(item *SparseExtent, offset int64) int {
			if offset < item.offset {
				return 1
			} else if offset == item.offset {
				return 0
			}
			return -1
		})

	if n < 1 || n > len(self.extents) {
		return nil, 0, 0, io.EOF
	}

	extent := self.extents[n-1]
	if extent.offset > offset || extent.offset+extent.total_size < offset {
		return nil, 0, 0, io.EOF
	}

	start, length, err = extent.getGrainForOffset(offset - extent.offset)
	return extent.reader, start, length, err
}

func (self *VMDKContext) ReadAt(buf []byte, offset int64) (int, error) {
	i := int64(0)
	buf_len := int64(len(buf))

	for i < buf_len {
		reader, start, available_length, err := self.getGrainForOffset(offset)
		if err != nil {
			return 0, err
		}

		to_read := buf_len - i
		if to_read > available_length {
			to_read = available_length
		}
		n, err := reader.ReadAt(buf[i:i+to_read], start)
		if err != nil && err != io.EOF {
			return int(i), err
		}

		i += int64(n)
	}

	return int(i), nil
}

func GetVMDKContext(
	reader io.ReaderAt, size int,
	opener func(filename string) (
		reader io.ReaderAt, closer func(), err error),
) (*VMDKContext, error) {
	profile := NewVMDKProfile()
	res := &VMDKContext{
		profile: profile,
		reader:  reader,
	}

	if size > 64*1024 {
		size = 64 * 1024
	}

	buf := make([]byte, size)
	n, err := reader.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}

	state := ""
	for _, line := range strings.Split(string(buf[:n]), "\n") {
		if StartExtentRegex.MatchString(line) {
			state = "Extents"
			continue
		}

		if state == "Extents" {
			match := ExtentRegex.FindStringSubmatch(line)
			if len(match) > 0 {
				extent_type := match[3]
				extent_filename := match[4]

				// Try to open the extent file.
				reader, closer, err := opener(extent_filename)
				if err != nil {
					return nil, err
				}

				switch extent_type {
				case "SPARSE":
					extent, err := GetSpaseExtent(reader)
					if err != nil {
						return nil, fmt.Errorf("While opening %v: %w",
							extent_filename, err)
					}

					extent.offset = res.total_size
					extent.closer = closer
					extent.filename = extent_filename

					res.total_size += extent.total_size

					res.extents = append(res.extents, extent)

				default:
					return nil, errors.New("Unsupported extent type " + extent_type)
				}

			} else {
				state = ""
			}
		}
	}

	return res, nil
}
