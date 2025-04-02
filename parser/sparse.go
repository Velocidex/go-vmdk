package parser

import (
	"errors"
	"fmt"
	"io"
)

type SparseExtent struct {
	profile *VMDKProfile
	reader  io.ReaderAt

	header *SparseExtentHeader

	// Size of grains in bytes
	grain_size int64

	// Coverage of each grain table in bytes
	grain_table_coverage int64

	gde_offset int64

	total_size int64

	// The offset in the logical image where this extent sits.
	offset   int64
	filename string

	closer func()
}

func (self *SparseExtent) Close() {
	if self.closer != nil {
		self.closer()
	}
}

func (self *SparseExtent) Debug() {
	fmt.Println(self.header.DebugString())
}

func (self *SparseExtent) TotalSize() int64 {
	return self.total_size
}

func (self *SparseExtent) VirtualOffset() int64 {
	return self.offset
}

func (self *SparseExtent) ReadAt(buf []byte, offset int64) (int, error) {
	start, available_length, err := self.getGrainForOffset(offset)
	if err != nil {
		return 0, nil
	}

	to_read := int64(len(buf))
	if to_read > available_length {
		to_read = available_length
	}

	return self.reader.ReadAt(buf[:to_read], start)
}

func (self *SparseExtent) getGrainForOffset(offset int64) (
	start, length int64, err error) {

	grain_table_number := offset / self.grain_table_coverage
	grain_directory_entry := ParseUint32(
		self.reader, self.gde_offset+4*grain_table_number)
	if grain_directory_entry == 0 {
		return 0, 0, io.EOF
	}

	grain_entry_number := (offset % self.grain_table_coverage) / self.grain_size
	grain_table_entry := ParseUint32(self.reader,
		int64(grain_directory_entry*SECTOR_SIZE)+4*grain_entry_number)

	grain_start := int64(grain_table_entry) * SECTOR_SIZE
	offset_within_grain := offset % self.grain_size

	return grain_start + offset_within_grain, self.grain_size - offset_within_grain, nil
}

func GetSparseExtent(reader io.ReaderAt) (*SparseExtent, error) {
	profile := NewVMDKProfile()
	res := &SparseExtent{
		profile: profile,
		reader:  reader,
		header:  profile.SparseExtentHeader(reader, 0),
	}

	if res.header.magicNumber() != SPARSE_MAGICNUMBER {
		return nil, errors.New("Invalid magic")
	}

	if res.header.version() != 1 {
		return nil, errors.New("Unsupported version")
	}

	if res.header.grainSize() < 8 {
		return nil, errors.New("Grain size invalid")
	}

	if res.header.numGTEsPerGT() != 512 {
		return nil, errors.New("numGTEsPerGT must be 512")
	}

	res.grain_size = int64(res.header.grainSize() * SECTOR_SIZE)
	res.grain_table_coverage = int64(res.header.numGTEsPerGT()) * res.grain_size
	res.gde_offset = int64(res.header.gdOffset() * SECTOR_SIZE)
	res.total_size = int64(res.header.capacity() * SECTOR_SIZE)

	return res, nil
}
