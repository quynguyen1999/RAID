# RAID Simulation in Go

This project simulates RAID levels 0, 1, 4, and 5 using regular files as physical disks.
It was implemented as part of an Operating Systems assignment to understand RAID performance, redundancy, parity logic, and block-based storage systems.

The program:

- Simulates 5 physical disks using .dat files
- Implements RAID0, RAID1, RAID4, and RAID5
- Uses XOR-based parity (RAID5 and RAID4)
- Benchmarks write/read performance (100 MB workload)
- Prints throughput and total execution times
- Demonstrates storage capacity differences across RAID levels

---

## How to Run

### Requirements
- Go 1.20 or newer
- Windows, macOS, or Linux
- No external libraries required (only Go standard library)

### Run the program
From the project root (same folder as main.go):

```bash
go run .
```

The program will automatically:

1. Run correctness tests for RAID0, RAID1, RAID4, and RAID5
2. Run a full benchmark (100 MB read + write) for all four RAID levels
3. Generate disk files such as:

```
raid0_disk0.dat … raid0_disk4.dat
raid1_disk0.dat … raid1_disk4.dat
raid4_disk0.dat … raid4_disk4.dat
raid5_disk0.dat … raid5_disk4.dat
bench_raid0_disk0.dat … bench_raid5_disk4.dat
```

---

## Benchmark Description

The benchmark writes a 100 MB dataset in 4 KB blocks, then reads it back.

For each RAID level, it reports:

- Total write time
- Total read time
- Per-block latency
- Throughput in MB/s

These results are used to compare performance against standard RAID expectations.

---

# Project Structure

```
RAID/
 ├── main.go        // Entry point: runs tests and benchmarks
 ├── raid.go        // RAID implementations and Disk abstraction
 ├── bench.go       // Benchmark engine
 ├── go.mod         // Go module metadata
 └── README.md      // Documentation
```

---

# Design Overview

## RAID Interface

All RAID levels implement the following interface:

```go
type RAID interface {
    Write(blockNum int, data []byte) error
    Read(blockNum int) ([]byte, error)
}
```

## Disk Abstraction

Each simulated disk is a regular file that supports fixed-size block operations.

Functions:
- WriteBlock(blockNum, data)
- ReadBlock(blockNum)
- Uses blockSize = 4096 bytes
- Seeks to offset = blockNum * blockSize
- Calls File.Sync() after writes to simulate real disk delay

### Disk file creation
Each RAID instance creates 5 disk files:

```
<prefix>_disk0.dat
<prefix>_disk1.dat
<prefix>_disk2.dat
<prefix>_disk3.dat
<prefix>_disk4.dat
```


# RAID Level Logic

## RAID0 (Striping)

- Splits data across all 5 disks.
- No redundancy.
- Maximum capacity and performance.

Mapping:

```go
diskIndex = blockNum % numDisks
diskBlock = blockNum / numDisks
```


## RAID1 (Mirroring)

- Writes the same block to all disks.
- Reads from disk 0 (simplified).
- Strong redundancy, lowest usable capacity.


## RAID4 (Striping + Dedicated Parity Disk)

- Disks 0–3 store data.
- Disk 4 stores parity (XOR of data blocks in each stripe).

Write procedure:

1. Read all data blocks in the stripe
2. Replace the updated block
3. Recompute parity using XOR
4. Write updated data and parity block

Parity disk can become a bottleneck.


## RAID5 (Striping + Distributed Parity)

- Parity rotates across all disks.
- Data occupies the remaining disks.

Parity disk per stripe:

```go
parityDisk = stripe % numDisks
```

Write procedure:

1. Read other data blocks in the stripe
2. Recompute parity
3. Write updated data and parity

Reduces bottleneck of RAID4 by rotating parity.

# Benchmark Logic (bench.go)

For each RAID level:

1. Generate a 4 KB random block
2. Compute number of blocks for 100 MB
3. Write all blocks sequentially
4. Read all blocks sequentially
5. Measure:
   - Total write time
   - Total read time
   - Time per block
   - Throughput in MB/s


# Dependencies
This project uses only Go standard library packages:
- os
- io
- fmt
- time
- bytes
- math/rand


