package unique

import (
	"fmt"
	"sync"
	"time"
)

// Character set for Base-62 encoding (10 digits + 26 upper + 26 lower = 62)
const base62Charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Custom Epoch: A fixed point in time (e.g., 2025-08-01) to keep the timestamp small.
const customEpoch = int64(1754006400000) // 2025-08-01 in milliseconds

// --- Transaction ID (Standard Snowflake) Config ---
// Target length is 12. Padding for leading zeros is now active to ensure this length is met.
const targetTxLength = 12

const (
	txSequenceBits = 12 // 4096 IDs per millisecond
	txWorkerIDBits = 10 // 1024 workers (0-1023)
	// The timestamp field uses the remaining 41 bits, providing about 69 years of life.
)

const (
	// Shift lengths for Transaction ID
	txWorkerIDShift  = txSequenceBits
	txTimestampShift = txSequenceBits + txWorkerIDBits // 12 + 10 = 22

	// Mask for sequence (max 4095)
	txSequenceMask = int64(-1) ^ (int64(-1) << txSequenceBits)
)

// --- User ID (Strict 8-Char Constraint + 1024 Workers) Config ---
// The 8-digit length (up to 50 bits) allows for maximum lifespan while keeping 1024 workers.
const targetUserLength = 8
const (
	userSequenceBits = 4  // 16 IDs per millisecond
	userWorkerIDBits = 10 // 1024 workers (0-1023) - REQUIRED BY USER
	// The timestamp field uses the remaining 36 bits, providing approx. 2,185 years of life.
)

const (
	// Shift lengths for User ID
	userWorkerIDShift  = userSequenceBits
	userTimestampShift = userSequenceBits + userWorkerIDBits // 4 + 10 = 14

	userSequenceMask = int64(-1) ^ (int64(-1) << userSequenceBits)
)

// IDGenerator encapsulates the state and configuration for generating unique IDs.
type IDGenerator struct {
	mu            sync.Mutex
	lastTimestamp int64
	sequenceNum   int64
	workerID      int64
}

// NewIDGen creates a new IDGenerator with the specified worker ID.
func NewIDGen(workerID int64) (*IDGenerator, error) {
	// Worker ID must satisfy the 10-bit constraint (0 to 1023) now required by BOTH Tx and User ID.
	maxWorkerID := int64(1<<txWorkerIDBits) - 1 // 1023
	if workerID < 0 || workerID > maxWorkerID {
		return nil, fmt.Errorf("worker ID must be between 0 and %d to satisfy the 1024 worker capacity constraint (10-bit allocation)", maxWorkerID)
	}

	return &IDGenerator{
		workerID: workerID,
	}, nil
}

// toBase62 converts a large 64-bit integer into a Base-62 string.
func toBase62(n int64) string {
	if n == 0 {
		return string(base62Charset[0])
	}

	res := make([]byte, 0)
	base := int64(len(base62Charset))
	for n > 0 {
		remainder := n % base
		res = append([]byte{base62Charset[remainder]}, res...)
		n /= base
	}
	return string(res)
}

// GeTransactionID generates a unique, distributed-safe, alphanumeric code.
// Padding is now active to ensure the resulting length is exactly 12 characters.
func (gen *IDGenerator) GeTransactionID() (string, error) {
	gen.mu.Lock()
	defer gen.mu.Unlock()

	// 1. Get current time in milliseconds since the custom epoch
	now := time.Now().UnixMilli() - customEpoch

	if now < gen.lastTimestamp {
		// Time went backwards, indicates system clock error.
		return "", fmt.Errorf("clock moved backwards; refusing to generate ID for %dms", gen.lastTimestamp-now)
	}

	if now == gen.lastTimestamp {
		// 2. Same millisecond, increment the sequence number
		gen.sequenceNum = (gen.sequenceNum + 1) & txSequenceMask
		if gen.sequenceNum == 0 {
			// Sequence overflow for this millisecond, wait for the next millisecond
			for now <= gen.lastTimestamp {
				now = time.Now().UnixMilli() - customEpoch
			}
		}
	} else {
		// 3. New millisecond, reset the sequence number
		gen.sequenceNum = 0
	}
	gen.lastTimestamp = now

	// 4. Assemble the 63-bit ID using bitwise shifts (the Snowflake pattern)
	uniqueID := (now << txTimestampShift) | (gen.workerID << txWorkerIDShift) | gen.sequenceNum

	// 5. Encode the unique ID to a Base-62 string.
	base62Code := toBase62(uniqueID)

	// 6. Pad the code to the target 12-digit length with '0's
	paddedCode := base62Code
	for len(paddedCode) < targetTxLength {
		paddedCode = "0" + paddedCode
	}

	return paddedCode, nil
}

// GetUserID generates a unique, distributed-safe, 8-character alphanumeric code for user IDs.
func (gen *IDGenerator) GetUserID() (string, error) {
	gen.mu.Lock()
	defer gen.mu.Unlock()

	// 1. Get current time in milliseconds since the custom epoch
	now := time.Now().UnixMilli() - customEpoch

	if now < gen.lastTimestamp {
		// Time went backwards, indicates system clock error.
		return "", fmt.Errorf("clock moved backwards; refusing to generate ID for %dms", gen.lastTimestamp-now)
	}

	if now == gen.lastTimestamp {
		// 2. Same millisecond, increment the sequence number (using the smaller mask)
		gen.sequenceNum = (gen.sequenceNum + 1) & userSequenceMask
		if gen.sequenceNum == 0 {
			// Sequence overflow for this millisecond, wait for the next millisecond
			for now <= gen.lastTimestamp {
				now = time.Now().UnixMilli() - customEpoch
			}
		}
	} else {
		// 3. New millisecond, reset the sequence number
		gen.sequenceNum = 0
	}
	gen.lastTimestamp = now

	// 4. Assemble the 50-bit ID (Optimized User Snowflake pattern)
	uniqueID := (now << userTimestampShift) | (gen.workerID << userWorkerIDShift) | gen.sequenceNum

	// 5. Encode the unique ID to a Base-62 string.
	base62Code := toBase62(uniqueID)

	// 6. Pad the code to the 8-digit length (only needed for very low time values near epoch start)
	paddedCode := base62Code
	for len(paddedCode) < targetUserLength {
		paddedCode = "0" + paddedCode
	}

	return paddedCode, nil
}
