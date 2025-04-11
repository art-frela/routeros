package routeros

import (
	"encoding/json"
	"time"
)

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
func (dt *DateTime) MarshalJSON() ([]byte, error) {
	if dt == nil {
		return nil, nil
	}

	return json.Marshal(dt.Format(time.DateTime))
}
