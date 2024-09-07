package parser

type ExtentStat struct {
	Type          string `json:"type"`
	VirtualOffset int64  `json:"VirtualOffset"`
	Size          int64  `json:"Size"`
	Filename      string `json:"Filename"`
}

type VMDKStats struct {
	TotalSize int64        `json:"TotalSize"`
	Extents   []ExtentStat `json:"Extents"`
}

func (self *SparseExtent) Stats() ExtentStat {
	return ExtentStat{
		Type:          "SPARSE",
		VirtualOffset: self.offset,
		Size:          self.total_size,
		Filename:      self.filename,
	}
}

func (self *VMDKContext) Stats() VMDKStats {
	res := VMDKStats{
		TotalSize: self.total_size,
	}

	for _, e := range self.extents {
		res.Extents = append(res.Extents, e.Stats())
	}

	return res
}
