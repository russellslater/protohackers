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

type testDispatcher struct {
	id    string
	roads []uint16
}

func (td *testDispatcher) ID() string {
	return td.id
}

func (td *testDispatcher) Roads() []uint16 {
	return td.roads
}

func (td *testDispatcher) SendTicket(t *ticketer.Ticket) {

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

func TestDetectSpeeding(t *testing.T) {
	tt := []struct {
		name         string
		o1           *ticketer.Observation
		o2           *ticketer.Observation
		wantSpeed    uint16
		wantSpeeding bool
	}{
		{
			name:         "Nil observations",
			o1:           nil,
			o2:           nil,
			wantSpeed:    0,
			wantSpeeding: false,
		},
		{
			name:         "First observation is nil",
			o1:           nil,
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			wantSpeed:    0,
			wantSpeeding: false,
		},
		{
			name:         "Second observation is nil",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:           nil,
			wantSpeed:    0,
			wantSpeeding: false,
		},
		{
			name:         "Zero distance and time between observations for same plate-road",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			wantSpeed:    0,
			wantSpeeding: false,
		},
		{
			name:         "Zero distance but some time between observations for same plate-road",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1001000},
			wantSpeed:    0,
			wantSpeeding: false,
		},
		{
			name:         "Zero time but some distance between observations for same plate-road",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 20, Plate: "HELLO", Timestamp: 1000000},
			wantSpeed:    0,
			wantSpeeding: false,
		},
		{
			name:         "Uncomparable observations - different plates, same road",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 20, Plate: "WORLD", Timestamp: 1000000},
			wantSpeed:    0,
			wantSpeeding: false,
		},
		{
			name:         "Uncomparable observations - same plates, different road",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{2, 100}, Mile: 20, Plate: "HELLO", Timestamp: 1000000},
			wantSpeed:    0,
			wantSpeeding: false,
		},
		{
			name:         "Speed below 100 mph speed limit",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 11, Plate: "HELLO", Timestamp: 1003600},
			wantSpeed:    1,
			wantSpeeding: false,
		},
		{
			name:         "Speed at 100 mph speed limit",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 100, Plate: "HELLO", Timestamp: 1003600},
			wantSpeed:    100,
			wantSpeeding: false,
		},
		{
			name:         "Speed just above 100 mph speed limit",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 100, Plate: "HELLO", Timestamp: 1003582},
			wantSpeed:    101,
			wantSpeeding: true,
		},
		{
			name:         "Speed well above 100 mph speed limit",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 10000, Plate: "HELLO", Timestamp: 1000001},
			wantSpeed:    20736,
			wantSpeeding: true,
		},
		{
			name:         "Going nowhere with a 0 mph speed limit",
			o1:           &ticketer.Observation{Road: &ticketer.Road{1, 0}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			o2:           &ticketer.Observation{Road: &ticketer.Road{1, 0}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			wantSpeed:    0,
			wantSpeeding: false,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			tm := ticketer.NewTicketManager()

			gotSpeed, gotSpeeding := tm.DetectSpeeding(tc.o1, tc.o2)
			is.Equal(gotSpeed, tc.wantSpeed)       // speed mismatch
			is.Equal(gotSpeeding, tc.wantSpeeding) // speeding detection mismatch
		})
	}
}

func TestDetectSpeedingInfractions(t *testing.T) {
	tt := []struct {
		name              string
		startObservations []*ticketer.Observation
		observation       *ticketer.Observation
		want              []*ticketer.Ticket
	}{
		{
			name:              "Empty observations",
			startObservations: nil,
			observation:       nil,
			want:              []*ticketer.Ticket{},
		},
		{
			name:              "First observation",
			startObservations: nil,
			observation:       &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			want:              []*ticketer.Ticket{},
		},
		{
			name: "First observation",
			startObservations: []*ticketer.Observation{
				{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			},
			observation: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 100, Plate: "HELLO", Timestamp: 1003400},
			want: []*ticketer.Ticket{
				{Plate: "HELLO", Road: 1, MileStart: 0, MileEnd: 100, TimestampStart: 1000000, TimestampEnd: 1003400, Speed: 10600},
			},
		},
		{
			name: "Ticket per infraction",
			startObservations: []*ticketer.Observation{
				{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
				{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
			},
			observation: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 100, Plate: "HELLO", Timestamp: 1003400},
			want: []*ticketer.Ticket{
				{Plate: "HELLO", Road: 1, MileStart: 0, MileEnd: 100, TimestampStart: 1000000, TimestampEnd: 1003400, Speed: 10600},
				{Plate: "HELLO", Road: 1, MileStart: 0, MileEnd: 100, TimestampStart: 1000000, TimestampEnd: 1003400, Speed: 10600},
			},
		},
		{
			name: "Final mile causes infraction",
			startObservations: []*ticketer.Observation{
				{Road: &ticketer.Road{1, 100}, Mile: 0, Plate: "HELLO", Timestamp: 1000000},
				{Road: &ticketer.Road{1, 100}, Mile: 50, Plate: "HELLO", Timestamp: 1001800},
			},
			observation: &ticketer.Observation{Road: &ticketer.Road{1, 100}, Mile: 51, Plate: "HELLO", Timestamp: 1001835},
			want: []*ticketer.Ticket{
				{Plate: "HELLO", Road: 1, MileStart: 50, MileEnd: 51, TimestampStart: 1001800, TimestampEnd: 1001835, Speed: 10300},
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			tm := ticketer.NewTicketManager()
			for _, o := range tc.startObservations {
				tm.Observe(o)
			}

			tickets := tm.DetectSpeedingInfractions(tc.observation)
			is.Equal(tickets, tc.want) // ticket mismatch
		})
	}
}

func TestAddDispatcher(t *testing.T) {
	tt := []struct {
		name                 string
		dispatchers          []ticketer.Dispatcher
		wantDispatcherCounts map[uint16]int
	}{
		{
			name:                 "Single dispatcher",
			dispatchers:          []ticketer.Dispatcher{&testDispatcher{id: "dispatcher_1", roads: []uint16{1, 2, 3}}},
			wantDispatcherCounts: map[uint16]int{1: 1, 2: 1, 3: 1},
		},
		{
			name: "Multiple dispatchers",
			dispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_1", roads: []uint16{1, 2, 3}},
				&testDispatcher{id: "dispatcher_2", roads: []uint16{2}},
				&testDispatcher{id: "dispatcher_3", roads: []uint16{2, 3}},
				&testDispatcher{id: "dispatcher_4", roads: []uint16{10}},
				&testDispatcher{id: "dispatcher_5", roads: []uint16{10, 11, 2}},
			},
			wantDispatcherCounts: map[uint16]int{1: 1, 2: 4, 3: 2, 10: 2, 11: 1},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			tm := ticketer.NewTicketManager()

			for _, d := range tc.dispatchers {
				tm.AddDispatcher(d)
			}

			for k, wantCount := range tc.wantDispatcherCounts {
				is.Equal(len(tm.Dispatchers[k]), wantCount) // dispatcher count for road mismatch
			}
		})
	}
}

func TestRemoveDispatcher(t *testing.T) {
	tt := []struct {
		name                 string
		startDispatchers     []ticketer.Dispatcher
		dispatchers          []ticketer.Dispatcher
		wantDispatcherCounts map[uint16]int
	}{
		{
			name:                 "Remove dispatcher that does not exist",
			dispatchers:          []ticketer.Dispatcher{&testDispatcher{id: "dispatcher_1", roads: []uint16{1, 2, 3}}},
			wantDispatcherCounts: map[uint16]int{1: 0, 2: 0, 3: 0},
		},
		{
			name:                 "Add then remove dispatcher",
			startDispatchers:     []ticketer.Dispatcher{&testDispatcher{id: "dispatcher_1", roads: []uint16{1, 2, 3}}},
			dispatchers:          []ticketer.Dispatcher{&testDispatcher{id: "dispatcher_1", roads: []uint16{1, 2, 3}}},
			wantDispatcherCounts: map[uint16]int{1: 0, 2: 0, 3: 0},
		},
		{
			name: "Add then remove multiple dispatchers",
			startDispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_1", roads: []uint16{1, 2, 3}},
				&testDispatcher{id: "dispatcher_2", roads: []uint16{2}},
				&testDispatcher{id: "dispatcher_3", roads: []uint16{2, 3}},
				&testDispatcher{id: "dispatcher_4", roads: []uint16{10}},
				&testDispatcher{id: "dispatcher_5", roads: []uint16{10, 11, 2}},
			},
			dispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_1", roads: []uint16{1, 2, 3}},
				&testDispatcher{id: "dispatcher_2", roads: []uint16{2}},
				&testDispatcher{id: "dispatcher_3", roads: []uint16{2, 3}},
				&testDispatcher{id: "dispatcher_4", roads: []uint16{10}},
				&testDispatcher{id: "dispatcher_5", roads: []uint16{10, 11, 2}},
			},
			wantDispatcherCounts: map[uint16]int{1: 0, 2: 0, 3: 0, 10: 0, 11: 0},
		},
		{
			name: "Add then remove some dispatchers",
			startDispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_1", roads: []uint16{1, 2, 3}},
				&testDispatcher{id: "dispatcher_2", roads: []uint16{2}},
				&testDispatcher{id: "dispatcher_3", roads: []uint16{2, 3}},
				&testDispatcher{id: "dispatcher_4", roads: []uint16{10}},
				&testDispatcher{id: "dispatcher_5", roads: []uint16{10, 11, 2}},
			},
			dispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_2", roads: []uint16{2}},
				&testDispatcher{id: "dispatcher_3", roads: []uint16{2, 3}},
				&testDispatcher{id: "dispatcher_5", roads: []uint16{10, 11, 2}},
			},
			wantDispatcherCounts: map[uint16]int{1: 1, 2: 1, 3: 1, 10: 1, 11: 0},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			tm := ticketer.NewTicketManager()

			for _, d := range tc.startDispatchers {
				tm.AddDispatcher(d)
			}

			for _, d := range tc.dispatchers {
				tm.RemoveDispatcher(d)
			}

			for k, wantCount := range tc.wantDispatcherCounts {
				is.Equal(len(tm.Dispatchers[k]), wantCount) // dispatcher count for road mismatch
			}
		})
	}
}
