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
			is.Equal(len(tm.Observations), tc.wantCount) // unexpected number of plate-road observations
		})
	}
}

func TestCalculateSpeed(t *testing.T) {
	tt := []struct {
		name string
		o1   *ticketer.Observation
		o2   *ticketer.Observation
		want uint16
	}{
		{
			name: "Nil observations",
			o1:   nil,
			o2:   nil,
			want: 0,
		},
		{
			name: "First observation is nil",
			o1:   nil,
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			want: 0,
		},
		{
			name: "Second observation is nil",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:   nil,
			want: 0,
		},
		{
			name: "Zero distance and time between observations for same plate-road",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			want: 0,
		},
		{
			name: "Zero distance but some time between observations for same plate-road",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1001000},
			want: 0,
		},
		{
			name: "Zero time but some distance between observations for same plate-road",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 20, Plate: "HELLO", Timestamp: 1000000},
			want: 0,
		},
		{
			name: "Uncomparable observations - different plates, same road",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 20, Plate: "WORLD", Timestamp: 1000000},
			want: 0,
		},
		{
			name: "Uncomparable observations - same plates, different road",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{2, 100}, Mile: 20, Plate: "HELLO", Timestamp: 1000000},
			want: 0,
		},
		{
			name: "Speed of 1 mph",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 11, Plate: "HELLO", Timestamp: 1003600},
			want: 1,
		},
		{
			name: "Speed of 5 mph",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 5, Plate: "HELLO", Timestamp: 1003600},
			want: 5,
		},
		{
			name: "Same speed of 5 mph when miles reversed",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 15, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1003600},
			want: 5,
		},
		{
			name: "Same speed of 5 mph when timestamps reversed",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1003600},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 15, Plate: "HELLO", Timestamp: 1000000},
			want: 5,
		},
		{
			name: "Same speed of 5 mph when observations reversed",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 15, Plate: "HELLO", Timestamp: 1003600},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			want: 5,
		},
		{
			name: "Speed of 200 mph",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 110, Plate: "HELLO", Timestamp: 1001800},
			want: 200,
		},
		{
			name: "Speed is rounded up when >= .5",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 100, Plate: "HELLO", Timestamp: 1003582},
			want: 101,
		},
		{
			name: "Speed is rounded down when < 0.5",
			o1:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			o2:   &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 100, Plate: "HELLO", Timestamp: 1003583},
			want: 100,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			tm := ticketer.NewTicketManager()

			got := tm.CalculateSpeed(tc.o1, tc.o2)
			is.Equal(got, tc.want) // speed did not match
		})
	}
}
