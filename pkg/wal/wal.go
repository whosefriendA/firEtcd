// Package wal provides a simplified Write-Ahead Log implementation.
package wal

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	ErrCorrupt  = errors.New("wal: log file is corrupt")
	ErrNotFound = errors.New("wal: file not found")
	ErrClosed   = errors.New("wal: log is closed")
)

// RecordType defines the type of record in the WAL.
type RecordType byte

const (
	// RecordTypeEntry is a standard log entry.
	RecordTypeEntry RecordType = 1
	// RecordTypeState is a metadata state change (e.g., term, vote).
	RecordTypeState RecordType = 2
	// RecordTypeSnapshot points to a snapshot file.
	RecordTypeSnapshot RecordType = 3
)

const (
	// recordHeaderSize is the size of the fixed-size part of a record header.
	// 4 bytes for length, 1 byte for type, 4 bytes for CRC32.
	recordHeaderSize = 4 + 1 + 4
)

// WAL represents a Write-Ahead Log.
type WAL struct {
	dir    string   // Directory where log files are stored.
	file   *os.File // Current open log file for appending.
	writer *bufio.Writer
	seq    uint64 // Sequence number of the current file.
}

// Create creates a new WAL in the given directory.
func Create(dir string) (*WAL, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return Open(dir)
}

// Open opens the WAL in the given directory.
func Open(dir string) (*WAL, error) {
	names, err := listLogFiles(dir)
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		// Create the first log file.
		return newLogFile(dir, 1)
	}

	// Open the last log file for appending.
	lastFile := names[len(names)-1]
	f, err := os.OpenFile(filepath.Join(dir, lastFile), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	var seq uint64
	fmt.Sscanf(lastFile, "%x.wal", &seq)

	return &WAL{
		dir:    dir,
		file:   f,
		writer: bufio.NewWriter(f),
		seq:    seq,
	}, nil
}

// newLogFile creates a new WAL segment file.
func newLogFile(dir string, seq uint64) (*WAL, error) {
	path := filepath.Join(dir, fmt.Sprintf("%016x.wal", seq))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{
		dir:    dir,
		file:   f,
		writer: bufio.NewWriter(f),
		seq:    seq,
	}, nil
}

// Write appends a new record to the log.
func (w *WAL) Write(recType RecordType, data []byte) error {
	if w.writer == nil {
		return ErrClosed
	}

	// 1. Create header
	header := make([]byte, recordHeaderSize)
	// Length of Type + Data + CRC
	binary.BigEndian.PutUint32(header[0:4], uint32(1+len(data)+4))
	header[4] = byte(recType)

	// 2. Calculate CRC
	crc := crc32.NewIEEE()
	crc.Write(header[4:5]) // Write type
	crc.Write(data)        // Write data
	binary.BigEndian.PutUint32(header[5:9], crc.Sum32())

	// 3. Write header and data
	if _, err := w.writer.Write(header); err != nil {
		return err
	}
	if _, err := w.writer.Write(data); err != nil {
		return err
	}

	return w.writer.Flush()
}

// ReadAll reads all records from all log files in sequence.
// It's used for recovery on startup.
func (w *WAL) ReadAll() (records [][]byte, types []RecordType, err error) {
	names, err := listLogFiles(w.dir)
	if err != nil {
		return nil, nil, err
	}

	for _, name := range names {
		path := filepath.Join(w.dir, name)
		f, err := os.Open(path)
		if err != nil {
			return nil, nil, err
		}
		defer f.Close()

		reader := bufio.NewReader(f)
		for {
			// 1. Read header
			header := make([]byte, recordHeaderSize)
			_, err := io.ReadFull(reader, header)
			if err == io.EOF {
				break // End of file, normal exit
			}
			if err != nil {
				return nil, nil, err
			}

			// 2. Decode header
			length := binary.BigEndian.Uint32(header[0:4])
			recType := RecordType(header[4])
			expectedCrc := binary.BigEndian.Uint32(header[5:9])

			// 3. Read data
			dataLen := length - (1 + 4) // Subtract type and CRC size
			data := make([]byte, dataLen)
			if _, err := io.ReadFull(reader, data); err != nil {
				return nil, nil, err
			}

			// 4. Verify CRC
			crc := crc32.NewIEEE()
			crc.Write([]byte{byte(recType)})
			crc.Write(data)
			if crc.Sum32() != expectedCrc {
				return nil, nil, ErrCorrupt
			}

			records = append(records, data)
			types = append(types, recType)
		}
	}
	return records, types, nil
}

// Truncate removes all log files up to and including the one containing lastIndex.
// This should be called after a snapshot is successfully saved.
func (w *WAL) Truncate(lastIndex uint64) error {
	names, err := listLogFiles(w.dir)
	if err != nil {
		return err
	}

	var seq uint64
	for _, name := range names {
		fmt.Sscanf(name, "%x.wal", &seq)
		if seq <= lastIndex {
			if err := os.Remove(filepath.Join(w.dir, name)); err != nil {
				// Log error but continue trying to delete others
			}
		}
	}
	return nil
}

// Close closes the currently open log file.
func (w *WAL) Close() error {
	if w.writer == nil {
		return nil // Already closed
	}
	if err := w.writer.Flush(); err != nil {
		return err
	}
	err := w.file.Close()
	w.writer = nil
	w.file = nil
	return err
}

// listLogFiles finds all .wal files in a directory and returns them sorted.
func listLogFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No directory means no files
		}
		return nil, err
	}

	var names []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".wal") {
			names = append(names, f.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
