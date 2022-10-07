package db_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/russellslater/protohackers/internal/db"
)

func TestSetAndGet(t *testing.T) {
	tt := []struct {
		name              string
		key               string
		wantInitialExists bool
		wantInitialValue  string
		value             string
		wantValue         string
	}{
		{
			name:              "Simple Key",
			key:               "foo",
			wantInitialExists: false,
			wantInitialValue:  "",
			value:             "bar",
			wantValue:         "bar",
		},
		{
			name:              "Empty Key",
			key:               "",
			wantInitialExists: false,
			wantInitialValue:  "",
			value:             "bar",
			wantValue:         "bar",
		},
		{
			name:              "Empty Value",
			key:               "foo",
			wantInitialExists: false,
			wantInitialValue:  "",
			value:             "",
			wantValue:         "",
		},
		{
			name:              "Empty Key & Value",
			key:               "",
			wantInitialExists: false,
			wantInitialValue:  "",
			value:             "",
			wantValue:         "",
		},
		{
			name:              "Default Nonoverridable Version",
			key:               "version",
			wantInitialExists: true,
			wantInitialValue:  "Ken's Key-Value Store 1.0",
			value:             "1.1",
			wantValue:         "Ken's Key-Value Store 1.0",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := db.NewUnusualDatabase()

			got, exists := db.Get(tc.key)

			is.Equal(exists, tc.wantInitialExists) // initial value existence mismatch
			is.Equal(got, tc.wantInitialValue)     // initial value mismatch

			db.Set(tc.key, tc.value)

			got, exists = db.Get(tc.key)

			is.True(exists)             // post-set value existence mismatch
			is.Equal(got, tc.wantValue) // post-set value mismatch
		})
	}
}
