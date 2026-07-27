package Models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// Vector is a pgvector-compatible float32 slice that implements the
// GORM driver.Valuer and sql.Scanner interfaces.
//
// GORM tag example:
//
//	Embedding Vector `gorm:"type:vector(1536);index:...,type:hnsw,opclass:vector_cosine_ops"`
type Vector []float32

// Value serialises the vector to the pgvector wire format "[x,y,z,...]".
func (v Vector) Value() (driver.Value, error) {
	if len(v) == 0 {
		return nil, nil
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 32))
	}
	sb.WriteByte(']')
	return sb.String(), nil
}

// Scan deserialises a pgvector "[x,y,z,...]" string back into the Vector slice.
func (v *Vector) Scan(src interface{}) error {
	var s string
	switch val := src.(type) {
	case string:
		s = val
	case []byte:
		s = string(val)
	case nil:
		*v = nil
		return nil
	default:
		return fmt.Errorf("vector: unsupported scan type %T", src)
	}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		*v = nil
		return nil
	}
	parts := strings.Split(s, ",")
	out := make(Vector, 0, len(parts))
	for _, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil {
			return fmt.Errorf("vector: parse error at %q: %w", p, err)
		}
		out = append(out, float32(f))
	}
	*v = out
	return nil
}
