package redis

import (
	"fmt"
)

const (
	TypeKey       = DataType("key")
	TypeList      = DataType("list")
	TypeSet       = DataType("set")
	TypeSortedSet = DataType("zset")
	TypeHash      = DataType("hash")
)

type DataType string

func (dt *DataType) UnmarshalTOML(src interface{}) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("cannot unmarshal %v", src)
	}
	switch v := DataType(s); v {
	case TypeKey, TypeList, TypeSet, TypeSortedSet, TypeHash:
		*dt = v
		return nil
	default:
		return fmt.Errorf("unknown Redis type: %s", s)
	}
}

func (dt DataType) String() string {
	return string(dt)
}
