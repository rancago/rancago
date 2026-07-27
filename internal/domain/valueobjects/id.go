package valueobjects

import (
	"fmt"
	"strconv"
)

type ID struct {
	uintVal  uint
	strVal   string
	useStr   bool
}

func NewIDUint(v uint) ID {
	return ID{uintVal: v, useStr: false, strVal: strconv.FormatUint(uint64(v), 10)}
}

func NewIDStr(v string) ID {
	return ID{strVal: v, useStr: true}
}

func (id ID) Uint() (uint, error) {
	if !id.useStr {
		return id.uintVal, nil
	}
	u, err := strconv.ParseUint(id.strVal, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("id is not uint: %w", err)
	}
	return uint(u), nil
}

func (id ID) String() string {
	if id.useStr {
		return id.strVal
	}
	return strconv.FormatUint(uint64(id.uintVal), 10)
}

func (id ID) IsString() bool { return id.useStr }
func (id ID) IsZero() bool {
	if id.useStr {
		return id.strVal == ""
	}
	return id.uintVal == 0
}
