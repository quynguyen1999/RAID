package main

import (
	"fmt"
	"io"
	"os"
)

const (
	blockSize = 4096
	numDisks  = 5
)

// RAID Interface

type RAID interface {
	Write(blockNum int, data []byte) error
	Read(blockNum int) ([]byte, error)
}

// Disk Layer

type Disk struct {
	f         *os.File
	blockSize int
}

// Create a disk file
func newDisk(filename string, blockSize int) (*Disk, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &Disk{f: f, blockSize: blockSize}, nil
}

func (d *Disk) WriteBlock(blockNum int, data []byte) error {
	if len(data) != d.blockSize {
		return fmt.Errorf("WriteBlock expected %d bytes, got %d", d.blockSize, len(data))
	}

	offset := int64(blockNum) * int64(d.blockSize)

	if _, err := d.f.Seek(offset, 0); err != nil {
		return err
	}
	if _, err := d.f.Write(data); err != nil {
		return err
	}

	return d.f.Sync()
}

func (d *Disk) ReadBlock(blockNum int) ([]byte, error) {
	buf := make([]byte, d.blockSize)
	offset := int64(blockNum) * int64(d.blockSize)

	if _, err := d.f.Seek(offset, 0); err != nil {
		return nil, err
	}

	n, err := d.f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Fill unwritten bytes with zeros
	for i := n; i < len(buf); i++ {
		buf[i] = 0
	}

	return buf, nil
}

func (d *Disk) Close() error {
	return d.f.Close()
}

// Create 5 disk files for a RAID level
func createDisks(prefix string, blockSize int) ([]*Disk, error) {
	disks := make([]*Disk, 0, numDisks)

	for i := 0; i < numDisks; i++ {
		filename := fmt.Sprintf("%s_disk%d.dat", prefix, i)
		_ = os.Remove(filename)

		d, err := newDisk(filename, blockSize)
		if err != nil {
			return nil, err
		}
		disks = append(disks, d)
	}

	return disks, nil
}

func xorBlocks(blocks ...[]byte) []byte {
	if len(blocks) == 0 {
		return nil
	}
	size := len(blocks[0])
	out := make([]byte, size)
	for _, b := range blocks {
		for i := 0; i < size; i++ {
			out[i] ^= b[i]
		}
	}
	return out
}

// RAID 0 (Striping)
type RAID0 struct {
	disks []*Disk
}

func NewRAID0(prefix string, blockSize int) (*RAID0, error) {
	disks, err := createDisks(prefix, blockSize)
	if err != nil {
		return nil, err
	}
	return &RAID0{disks: disks}, nil
}

func (r *RAID0) Write(blockNum int, data []byte) error {
	diskIndex := blockNum % numDisks
	diskBlock := blockNum / numDisks
	return r.disks[diskIndex].WriteBlock(diskBlock, data)
}

func (r *RAID0) Read(blockNum int) ([]byte, error) {
	diskIndex := blockNum % numDisks
	diskBlock := blockNum / numDisks
	return r.disks[diskIndex].ReadBlock(diskBlock)
}

// RAID 1 (Mirroring)

type RAID1 struct {
	disks []*Disk
}

func NewRAID1(prefix string, blockSize int) (*RAID1, error) {
	disks, err := createDisks(prefix, blockSize)
	if err != nil {
		return nil, err
	}
	return &RAID1{disks: disks}, nil
}

func (r *RAID1) Write(blockNum int, data []byte) error {
	for _, d := range r.disks {
		if err := d.WriteBlock(blockNum, data); err != nil {
			return err
		}
	}
	return nil
}

func (r *RAID1) Read(blockNum int) ([]byte, error) {
	return r.disks[0].ReadBlock(blockNum)
}

// RAID 4 (Striping + dedicated parity)

type RAID4 struct {
	disks []*Disk
}

func NewRAID4(prefix string, blockSize int) (*RAID4, error) {
	disks, err := createDisks(prefix, blockSize)
	if err != nil {
		return nil, err
	}
	return &RAID4{disks: disks}, nil
}

func (r *RAID4) Write(blockNum int, data []byte) error {
	dataDisks := 4
	stripe := blockNum / dataDisks
	offset := blockNum % dataDisks // which data disk (0..3) in this stripe

	// read all data blocks in this stripe
	dataBlocks := make([][]byte, dataDisks)
	for i := 0; i < dataDisks; i++ {
		if i == offset {
			dataBlocks[i] = data
		} else {
			b, err := r.disks[i].ReadBlock(stripe)
			if err != nil {
				return err
			}
			dataBlocks[i] = b
		}
	}

	// compute parity over the 4 data blocks
	parity := xorBlocks(dataBlocks...)

	// write the updated data block
	if err := r.disks[offset].WriteBlock(stripe, data); err != nil {
		return err
	}

	// write parity to disk 4
	if err := r.disks[4].WriteBlock(stripe, parity); err != nil {
		return err
	}

	return nil
}

func (r *RAID4) Read(blockNum int) ([]byte, error) {
	dataDisks := 4
	stripe := blockNum / dataDisks
	offset := blockNum % dataDisks
	return r.disks[offset].ReadBlock(stripe)
}

// RAID 5 (Striping + distributed parity)

type RAID5 struct {
	disks []*Disk
}

func NewRAID5(prefix string, blockSize int) (*RAID5, error) {
	disks, err := createDisks(prefix, blockSize)
	if err != nil {
		return nil, err
	}
	return &RAID5{disks: disks}, nil
}

// layout returns mapping info for a logical block:
// which physical data disk, which block on disk, which parity disk
func (r *RAID5) layout(blockNum int) (dataDiskIdx int, diskBlock int, parityDiskIdx int, dataIdx int, stripe int) {
	n := numDisks
	dataPerStripe := n - 1

	stripe = blockNum / dataPerStripe
	dataIdx = blockNum % dataPerStripe // 0..3: index among data blocks
	parityDiskIdx = stripe % n         // parity rotates

	dataDiskIdx = dataIdx
	if dataDiskIdx >= parityDiskIdx {
		dataDiskIdx++ // skip parity disk
	}
	diskBlock = stripe
	return
}

func (r *RAID5) Write(blockNum int, data []byte) error {
	n := numDisks
	dataPerStripe := n - 1

	dataDiskIdx, diskBlock, parityDiskIdx, dataIdx, stripe := r.layout(blockNum)

	// gather all data blocks in the stripe
	dataBlocks := make([][]byte, dataPerStripe)
	for i := 0; i < dataPerStripe; i++ {
		if i == dataIdx {
			dataBlocks[i] = data
		} else {
			// map logical data index i to physical disk index
			d := i
			if d >= parityDiskIdx {
				d++
			}
			b, err := r.disks[d].ReadBlock(stripe)
			if err != nil {
				return err
			}
			dataBlocks[i] = b
		}
	}

	parity := xorBlocks(dataBlocks...)

	// write data block
	if err := r.disks[dataDiskIdx].WriteBlock(diskBlock, data); err != nil {
		return err
	}

	// write parity block
	if err := r.disks[parityDiskIdx].WriteBlock(diskBlock, parity); err != nil {
		return err
	}

	return nil
}

func (r *RAID5) Read(blockNum int) ([]byte, error) {
	dataDiskIdx, diskBlock, _, _, _ := r.layout(blockNum)
	return r.disks[dataDiskIdx].ReadBlock(diskBlock)
}

// RAID Factory (Choose level 0,1,4,5)

func NewRAID(level string, prefix string, blockSize int) (RAID, error) {
	switch level {
	case "0":
		return NewRAID0(prefix, blockSize)
	case "1":
		return NewRAID1(prefix, blockSize)
	case "4":
		return NewRAID4(prefix, blockSize)
	case "5":
		return NewRAID5(prefix, blockSize)
	default:
		return nil, fmt.Errorf("unsupported RAID level: %s", level)
	}
}
