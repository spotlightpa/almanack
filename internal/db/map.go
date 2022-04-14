package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Map map[string]any

// Value implements the driver.Valuer interface.
func (m Map) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	return b, err
}

// Scan implements the sql.Scanner interface.
func (m *Map) Scan(value any) error {
	dbMap := make(map[string]any)
	if value == nil {
		*m = dbMap
		return nil
	}
	buf, ok := value.([]byte)
	if !ok {
		return errors.New("canot parse to bytes")
	}
	if err := json.Unmarshal(buf, &dbMap); err != nil {
		return err
	}
	*m = dbMap
	return nil
}
