package fit

import (
	"encoding/binary"
	"io"
)

// FitEncoder encodes FIT activity files using streaming writes with optimized CRC calculation
type FitEncoder struct {
	w          io.WriteSeeker
	crc        uint16
	dataSize   int
	headerSize int
	startPos   int64 // position after header
}

// NewFitEncoder creates a new streaming FIT encoder
func NewFitEncoder(w io.WriteSeeker) (*FitEncoder, error) {
	encoder := &FitEncoder{
		w:          w,
		crc:        0,
		dataSize:   0,
		headerSize: 14, // Standard header size with CRC
	}

	// Get current position for header
	var err error
	if encoder.startPos, err = w.Seek(0, io.SeekCurrent); err != nil {
		return nil, err
	}

	// Write header placeholder
	header := []byte{
		14,            // Header size
		0x10,          // Protocol version
		0x00, 0x2D,    // Profile version (little endian 45)
		0x00, 0x00, 0x00, 0x00, // Data size (4 bytes, will be updated later)
		'.', 'F', 'I', 'T', // ".FIT" data type
		0x00, 0x00, // Header CRC (will be calculated later)
	}

	// Write header and calculate CRC
	if _, err := w.Write(header); err != nil {
		return nil, err
	}
	encoder.updateCRC(header)

	return encoder, nil
}

// updateCRC calculates CRC-16 checksum without hash/crc16 dependency
func (e *FitEncoder) updateCRC(data []byte) {
	crcTable := [...]uint16{
		0x0000, 0xCC01, 0xD801, 0x1400, 0xF001, 0x3C00, 0x2800, 0xE401,
		0xA001, 0x6C00, 0x7800, 0xB401, 0x5000, 0x9C01, 0x8801, 0x4400,
	}

	currentCRC := e.crc
	for _, b := range data {
		// Compute checksum of lower four bits
		tmp := crcTable[currentCRC&0xF]
		currentCRC = (currentCRC >> 4) & 0x0FFF
		currentCRC = currentCRC ^ tmp ^ crcTable[b&0xF]

		// Compute checksum of upper four bits
		tmp = crcTable[currentCRC&0xF]
		currentCRC = (currentCRC >> 4) & 0x0FFF
		currentCRC = currentCRC ^ tmp ^ crcTable[(b>>4)&0xF]
	}
	e.crc = currentCRC
}

// Write writes activity data in chunks
func (e *FitEncoder) Write(p []byte) (int, error) {
	n, err := e.w.Write(p)
	if err != nil {
		return n, err
	}

	e.updateCRC(p)
	e.dataSize += n
	return n, nil
}

// Close finalizes the FIT file
func (e *FitEncoder) Close() error {
	// Save current position
	currentPos, err := e.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	// Update data size in header
	if _, err := e.w.Seek(e.startPos+4, io.SeekStart); err != nil {
		return err
	}
	dataSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(dataSizeBytes, uint32(e.dataSize))
	if _, err := e.w.Write(dataSizeBytes); err != nil {
		return err
	}

	// Recalculate header CRC with original data
	header := []byte{
		14,            // Header size
		0x10,          // Protocol version
		0x00, 0x2D,    // Profile version
		dataSizeBytes[0], dataSizeBytes[1], dataSizeBytes[2], dataSizeBytes[3],
		'.', 'F', 'I', 'T', // ".FIT" data type
	}
	
	// Calculate header CRC with clean state
	e.crc = 0
	e.updateCRC(header)
	headerCRC := e.crc

	// Update header CRC
	if _, err := e.w.Seek(e.startPos+12, io.SeekStart); err != nil {
		return err
	}
	crcBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(crcBytes, headerCRC)
	if _, err := e.w.Write(crcBytes); err != nil {
		return err
	}

	// Write final file CRC
	if _, err := e.w.Seek(currentPos, io.SeekStart); err != nil {
		return err
	}
	fileCRCBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(fileCRCBytes, e.crc)
	_, err = e.w.Write(fileCRCBytes)
	return err
}
