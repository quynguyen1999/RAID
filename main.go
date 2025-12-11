package main

import (
	"bytes"
	"fmt"
)

func main() {
	// Correctness Tests (write + read)

	raid0, err := NewRAID("0", "raid0_test", blockSize)
	if err != nil {
		panic(err)
	}
	testRAID("RAID0", raid0)

	raid1, err := NewRAID("1", "raid1_test", blockSize)
	if err != nil {
		panic(err)
	}
	testRAID("RAID1", raid1)

	raid4, err := NewRAID("4", "raid4_test", blockSize)
	if err != nil {
		panic(err)
	}
	testRAID("RAID4", raid4)

	raid5, err := NewRAID("5", "raid5_test", blockSize)
	if err != nil {
		panic(err)
	}
	testRAID("RAID5", raid5)

	// Benchmarks (100 MB write + read each)

	fmt.Println()
	fmt.Println("=========== BENCHMARKS ===========")

	for _, level := range []string{"0", "1", "4", "5"} {
		benchmark(level)
	}
}

// testRAID writes one block and reads it back to verify correctness.
func testRAID(name string, r RAID) {
	data := make([]byte, blockSize)
	copy(data, []byte("hello from "+name))

	blockNum := 0 // write at beginning for easy inspection

	if err := r.Write(blockNum, data); err != nil {
		panic(err)
	}

	out, err := r.Read(blockNum)
	if err != nil {
		panic(err)
	}

	trimmed := bytes.TrimRight(out, "\x00")
	fmt.Printf("%s read: %q\n", name, string(trimmed))
}
