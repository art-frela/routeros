package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTime_UnmarshalJSON(t *testing.T) {
	victoryDay, _ := time.Parse(time.RFC3339, "1945-05-09T09:00:00Z")
	tests := []struct {
		name         string
		data         []byte
		wantErr      bool
		wantDateTime time.Time
	}{
		{
			name:         "test.1 ok - 9 May 1945",
			data:         []byte(`"1945-05-09 09:00:00"`),
			wantErr:      false,
			wantDateTime: victoryDay,
		},
		{
			name:         "test.2 err - empty",
			data:         []byte(`""`),
			wantErr:      true,
			wantDateTime: time.Time{},
		},
		{
			name:         "test.3 err - no format",
			data:         []byte(`"1999-01-01T00:00:00+07:00"`),
			wantErr:      true,
			wantDateTime: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &DateTime{}
			err := dt.UnmarshalJSON(tt.data)

			if tt.wantErr {
				assert.Error(t, err, "DateTime.UnmarshalJSON() error")
			} else {
				assert.NoError(t, err, "DateTime.UnmarshalJSON() error")
			}

			assert.Equal(t, tt.wantDateTime, dt.Time)
		})
	}
}
