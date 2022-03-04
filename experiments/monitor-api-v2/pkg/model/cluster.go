package model

type NodeResource struct {
	CpuTotal int64
	MemTotal int64
	GpuTotal int
	// GpuMemTotal in bytes
	GpuMemTotal int64
	CpuFree     int64
	MemFree     int64
	// UpperBound in bytes
	CpuUB int64
	MemUB int64
	/* Available GPU calculate */
	// Total GPU count - Pods using nvidia.com/gpu
	GpuFreeCount int
}

func (this *NodeResource) DeepCopy() *NodeResource {
	return &NodeResource{
		CpuTotal:     this.CpuTotal,
		MemTotal:     this.MemTotal,
		GpuTotal:     this.GpuTotal,
		GpuMemTotal:  this.GpuMemTotal,
		CpuFree:      this.CpuFree,
		MemFree:      this.MemFree,
		CpuUB:        this.CpuUB,
		MemUB:        this.MemUB,
		GpuFreeCount: this.GpuFreeCount,
	}
}
