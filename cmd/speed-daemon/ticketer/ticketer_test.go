package ticketer_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/russellslater/protohackers/cmd/speed-daemon/ticketer"
)

type testObservation struct {
	ob        *ticketer.Observation
	wantMatch bool
}

func TestObserve(t *testing.T) {
	tt := []struct {
		name         string
		observations []testObservation
		wantCount    int
	}{
		{
			name:         "Zero observations",
			observations: []testObservation{{ob: nil, wantMatch: false}},
			wantCount:    0,
		},
		{
			name:         "One observed plate-road combination",
			observations: []testObservation{{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000}, wantMatch: false}},
			wantCount:    1,
		},
		{
			name: "Two observed plate-road combinations",
			observations: []testObservation{
				{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000}, wantMatch: false},
				{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 20, Plate: "WORLD", Timestamp: 1040000}, wantMatch: false},
			},
			wantCount: 2,
		},
		{
			name: "One observed plate-road combination with one match",
			observations: []testObservation{
				{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000}, wantMatch: false},
				{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 20, Plate: "HELLO", Timestamp: 1040000}, wantMatch: true},
			},
			wantCount: 1,
		},
		{
			name: "Four observed plate-road combinations with zero matches",
			observations: []testObservation{
				{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000}, wantMatch: false},
				{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 20, Plate: "WORLD", Timestamp: 1040000}, wantMatch: false},
				{ob: &ticketer.Observation{Road: &ticketer.Road{2, 100}, Mile: 30, Plate: "HELLO", Timestamp: 1040000}, wantMatch: false},
				{ob: &ticketer.Observation{Road: &ticketer.Road{2, 100}, Mile: 40, Plate: "WORLD", Timestamp: 1040000}, wantMatch: false},
			},
			wantCount: 4,
		},
		{
			name: "Four observed plate-road combinations with two matches",
			observations: []testObservation{
				{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000}, wantMatch: false},
				{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 20, Plate: "WORLD", Timestamp: 1040000}, wantMatch: false},
				{ob: &ticketer.Observation{Road: &ticketer.Road{2, 100}, Mile: 30, Plate: "HELLO", Timestamp: 1040000}, wantMatch: false},
				{ob: &ticketer.Observation{Road: &ticketer.Road{2, 100}, Mile: 40, Plate: "WORLD", Timestamp: 1040000}, wantMatch: false},
				{ob: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 15, Plate: "HELLO", Timestamp: 1003000}, wantMatch: true},
				{ob: &ticketer.Observation{Road: &ticketer.Road{2, 100}, Mile: 45, Plate: "WORLD", Timestamp: 1045000}, wantMatch: true},
			},
			wantCount: 4,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			tm := ticketer.NewTicketManager()

			for _, o := range tc.observations {
				got := tm.Observe(o.ob)
				is.Equal(got, o.wantMatch) // match mismatch
			}
			is.Equal(tc.wantCount, len(tm.Observations)) // unexpected number of plate-road observations
		})
	}
}
