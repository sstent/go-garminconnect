package fit

import (
	"bytes"
	"encoding/binary"
	"time"
)

// FitBaseType represents FIT base type definitions
type FitBaseType struct {
	ID      int
	Name    string
	Size    int
	Invalid uint64
	Field   byte
}

// Base types definitions
var (
	FitEnum = FitBaseType{0, "enum", 1, 0xFF, 0x00}
	FitUint8 = FitBaseType{2, "uint8", 1, 0xFF, 0x02}
	FitUint16 = FitBaseType{4, "uint16", 2, 0xFFFF, 0x84}
	FitUint32 = FitBaseType{6, "uint32", 4, 0xFFFFFFFF, 0x86}
	FitString = FitBaseType{7, "string", 1, 0x00, 0x07}
	FitFloat32 = FitBaseType{8, "float32", 4, 0xFFFFFFFF, 0x88}
	FitByte = FitBaseType{13, "byte", 1, 0xFF, 0x0D}
)

// FitEncoder encodes FIT activity files
type FitEncoder struct {
	buf bytes.Buffer
	headerSize int
	activityDefined bool
}

const (
	FitHeaderSize = 12
	FileTypeActivity = 4
	GarminEpochOffset = 631065600 // UTC 00:00 Dec 31 1989
)

// NewFitEncoder creates a new FIT encoder
func NewFitEncoder() *FitEncoder {
	e := &FitEncoder{headerSize: FitHeaderSize}
	e.writeHeader(0) // Initial header with 0 data size
	return e
}

// writeHeader writes the FIT file header
func (e *FitEncoder) writeHeader(dataSize int) {
	e.buf.Reset()
	header := []byte{
		byte(e.headerSize), // Header size
		16,                 // Protocol version
		0, 0, 0, 108,       // Profile version (108.0)
	}
	e.buf.Write(header)

	// Write data size (4 bytes, little-endian)
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(dataSize))
	e.buf.Write(sizeBytes)

	// Write file type signature
	e.buf.Write([]byte(".FIT"))
}

// AddActivity adds activity data to the FIT file
func (e *FitEncoder) AddActivity(activity FitActivity) error {
	// TODO: Implement activity message encoding
	return nil
}

// Encode returns the encoded FIT file bytes
func (e *FitEncoder) Encode() ([]byte, error) {
	dataSize := e.buf.Len() - e.headerSize
	e.writeHeader(dataSize)
	e.writeCRC()
	return e.buf.Bytes(), nil
}

// writeCRC calculates and appends the FIT CRC
func (e *FitEncoder) writeCRC() {
	data := e.buf.Bytes()
	crc := uint16(0)
	for _, b := range data {
		crc = calcCRC(crc, b)
	}
	crcBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(crcBytes, crc)
	e.buf.Write(crcBytes)
}

// calcCRC calculates FIT CRC
func calcCRC(crc uint16, byteVal byte) uint16 {
	table := [...]uint16{
		0x0000, 0xCC01, 0xD801, 0x1400, 0xF001, 0x3C00, 0x2800, 0xE401,
		0xA001, 0x6C00, 0x7800, 0xB401, 0x5000, 0x9C01, 0x8801, 0x4400,
	}
	tmp := table[crc & 0xF]
	crc = (crc >> 4) & 0x0FFF
	crc = crc ^ tmp ^ table[byteVal & 0xF]
	
	tmp = table[crc & 0xF]
	crc = (crc >> 4) & 0x0FFF
	return crc ^ tmp ^ table[(byteVal >> 4) & 0xF]
}

// timestamp converts Go time to FIT timestamp
func timestamp(t time.Time) uint32 {
	return uint32(t.Unix() - GarminEpochOffset)
}

// FitActivity represents basic activity data for FIT encoding
type FitActivity struct {
	Name      string
	Type      string
	StartTime time.Time
	Duration  time.Duration
	Distance  float32 // in meters
}
