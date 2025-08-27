package fit

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	headerSize    = 12
	protocolMajor = 2
)

// FileHeader represents the header of a FIT file
type FileHeader struct {
	Size      uint8
	Protocol  uint8
	Profile   [4]byte
	DataSize  uint32
	Signature [4]byte
}

// Activity represents activity data from a FIT file
type Activity struct {
	Type          string
	StartTime     int64
	TotalDistance float64
	Duration      float64
}

// Decoder parses FIT files
type Decoder struct {
	r io.Reader
}

// NewDecoder creates a new FIT decoder
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Parse decodes the FIT file and returns the activity data
func (d *Decoder) Parse() (*Activity, error) {
	var header FileHeader
	if err := binary.Read(d.r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	// Validate header
	if header.Protocol != protocolMajor {
		return nil, errors.New("unsupported FIT protocol version")
	}

	// For simplicity, we'll just extract basic activity data
	activity := &Activity{}

	// Skip to activity record (simplified for example)
	// In a real implementation, we would parse the file structure properly
	for {
		var recordHeader uint8
		if err := binary.Read(d.r, binary.LittleEndian, &recordHeader); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if recordHeader == 0x21 { // Activity record header (example value)
			var record struct {
				Type          uint8
				StartTime     int64
				TotalDistance float32
				Duration      uint32
			}
			if err := binary.Read(d.r, binary.LittleEndian, &record); err != nil {
				return nil, err
			}

			activity.Type = activityType(record.Type)
			activity.StartTime = record.StartTime
			activity.TotalDistance = float64(record.TotalDistance)
			activity.Duration = float64(record.Duration)
			break
		}
	}

	return activity, nil
}

func activityType(t uint8) string {
	switch t {
	case 1:
		return "Running"
	case 2:
		return "Cycling"
	case 3:
		return "Swimming"
	default:
		return "Unknown"
	}
}

// ReadFile reads and parses a FIT file
func ReadFile(path string) (*Activity, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return NewDecoder(file).Parse()
}
