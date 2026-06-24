package types

import (
	"encoding/json"
	"time"
)

// Error represents a RouterOS API docs error response.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type Error struct {
	Detail  string `json:"detail,omitempty"`
	Error   int    `json:"error"`
	Message string `json:"message"`
}

// DateTime is a custom time type that marshals/unmarshals
// RouterOS datetime strings in "2006-01-02 15:04:05" format.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type DateTime struct {
	time.Time
}

func (dt *DateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return err
	}

	dt.Time = t

	return nil
}

// MarshalJSON преобразует CustomTime обратно в JSON.
func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(dt.Time.Format(time.DateTime))
}
