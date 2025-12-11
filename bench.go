package main

import (
	"fmt"
	"math/rand"
	"time"
)

const totalMB = 100 // size of benchmark dataset

// benchmark runs a write+read benchmark for one RAID level.
// writes totalMB MB in blockSize chunks, reads it back,
// and prints total + per-block times and throughput.
func benchmark(level string) {
	fmt.Printf("Benchmark RAID level %s\n", level)

	// Create a RAID instance for this level
	prefix := "bench_raid" + level
	r, err := NewRAID(level, prefix, blockSize)
	if err != nil {
		panic(err)
	}

	// Total bytes and blocks
	totalBytes := totalMB * 1024 * 1024
	numBlocks := totalBytes / blockSize

	fmt.Printf("Total size: %d MB, block size: %d bytes, blocks: %d\n",
		totalMB, blockSize, numBlocks)

	// Use a single block buffer for all writes
	block := make([]byte, blockSize)
	rand.Seed(42)
	if _, err := rand.Read(block); err != nil {
		panic(err)
	}

	// WRITE benchmark
	startWrite := time.Now()

	for i := 0; i < numBlocks; i++ {
		if err := r.Write(i, block); err != nil {
			panic(fmt.Sprintf("write error at block %d: %v", i, err))
		}
	}

	writeDur := time.Since(startWrite)
	writePerBlock := writeDur / time.Duration(numBlocks)

	// READ benchmark
	startRead := time.Now()

	for i := 0; i < numBlocks; i++ {
		if _, err := r.Read(i); err != nil {
			panic(fmt.Sprintf("read error at block %d: %v", i, err))
		}
	}

	readDur := time.Since(startRead)
	readPerBlock := readDur / time.Duration(numBlocks)

	// Report
	bytesMB := float64(totalBytes) / (1024 * 1024)
	writeMBps := bytesMB / writeDur.Seconds()
	readMBps := bytesMB / readDur.Seconds()

	fmt.Printf("Write: total = %v, per block = %v, throughput = %.2f MB/s\n",
		writeDur, writePerBlock, writeMBps)
	fmt.Printf("Read : total = %v, per block = %v, throughput = %.2f MB/s\n\n",
		readDur, readPerBlock, readMBps)
}
