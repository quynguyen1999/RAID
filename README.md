#All simulated disk files are written in the current directory with names like:
raid0_disk0.dat … raid0_disk4.dat
raid1_disk0.dat … raid1_disk4.dat
raid4_disk0.dat … raid4_disk4.dat
raid5_disk0.dat … raid5_disk4.dat


#Design
Overall Structure
raid.go
Defines the RAID interface:
type RAID interface {
    Write(blockNum int, data []byte) error
    Read(blockNum int) ([]byte, error)
}


#Implements a Disk type that wraps a file and supports:
WriteBlock(blockNum int, data []byte)
ReadBlock(blockNum int) ([]byte, error)
Uses fixed blockSize = 4096 bytes and seeks to blockNum * blockSize.
Calls Sync() (fsync) after writes to simulate real disk delay.
Provides createDisks(prefix, blockSize) to create 5 disk files per RAID instance.

#Implements:
RAID0 (striping)
RAID1 (mirroring)
RAID4 (striping + dedicated parity disk)
RAID5 (striping + distributed parity)

#Includes a factory:
func NewRAID(level, prefix string, blockSize int) (RAID, error)
which returns the appropriate RAID implementation.

#RAID Level Logic
RAID0 (Striping)
Uses all 5 disks for data.
Mapping:
diskIndex = blockNum % numDisks
diskBlock = blockNum / numDisks
No redundancy, maximum usable capacity.

RAID1 (Mirroring)
Writes the same block to all disks.
Reads from the first disk (simplified model, no read load-balancing).
Usable capacity is one disk; strongest redundancy.
RAID4 (Striping + Dedicated Parity)
Disks 0–3: data, disk 4: parity.
For each stripe:
Reads all 4 data blocks, replaces one with the new data, recomputes XOR parity, writes data + parity.
Reads go directly to the appropriate data disk; parity is unused unless a failure is simulated.

RAID5 (Striping + Distributed Parity)
Parity disk rotates per stripe:
parityDisk = stripe % numDisks
Data blocks occupy all other disks in that stripe.
On write:
Reads all data blocks in the stripe (except the one being updated).
Recomputes parity with XOR.
Writes data to its data disk and parity to the parity disk.

#Benchmark Design
Benchmark code is in bench.go, and is invoked from main.go.
For each RAID level ("0", "1", "4", "5"):
Create a new RAID instance with prefix bench_raid<level>.
Set:
totalMB = 100
blockSize = 4096
numBlocks = (totalMB * 1024 * 1024) / blockSize

Write phase
Use a fixed 4 KB buffer (filled with random data once).
Call Write(i, block) for block i = 0..numBlocks-1.
Measure total write time and per-block time.

Read phase
Call Read(i) for each logical block.
Measure total read time and per-block time.

#Print:
Total time
Per-block time
Throughput in MB/s
These numbers are later compared against textbook expectations for RAID performance.

#Dependencies
Language: Go
Libraries: only Go standard library
os, io, fmt, time, math/rand, bytes