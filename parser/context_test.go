package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sebdah/goldie"
)

type MockExtent struct {
	*SparseExtent

	buf []byte
}

func (self *MockExtent) ReadAt(buf []byte, offset int64) (int, error) {
	for i := 0; i < len(buf); i++ {
		buf[i] = self.buf[i+int(offset)]
	}

	return len(buf), nil
}

func makeData(offset, length int) string {
	res := ""
	for len(res) < length {
		res += fmt.Sprintf(" % 4d", offset+len(res))
	}

	return res
}

func NewMockExtent(offset, total_size int64) Extent {
	return &MockExtent{
		SparseExtent: &SparseExtent{
			offset:     offset,
			total_size: total_size,
		},
		buf: []byte(makeData(int(offset), int(total_size))),
	}
}

func TestFindExtent(t *testing.T) {
	res := &VMDKContext{
		total_size: 350,
		extents: []Extent{
			NewMockExtent(0, 100),
			NewMockExtent(100, 100),
			// Gap
			NewMockExtent(300, 50),
		},
	}

	res.normalizeExtents()
	var golden []string

	for _, offset := range []int64{0, 5, 95, 210, 290, 340} {
		buf := make([]byte, 20)

		extent, err := res.getExtentForOffset(offset)
		if err != nil {
			golden = append(golden,
				fmt.Sprintf("err for %v %v\n", offset, err))
		} else {
			golden = append(golden,
				fmt.Sprintf("extent for %v %v, err %v\n",
					offset, extent.Stats(), err))
		}

		n, err := res.ReadAt(buf, offset)
		golden = append(golden,
			fmt.Sprintf("Reading %v (%v) : %v (%v)\n", offset, n,
				string(buf[:n]), err))
	}

	goldie.Assert(t, "TestFindExtent", []byte(strings.Join(golden, "\n")))
}
