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
	id              string
	roads           []ticketer.RoadID
	sentTicketCount int
	lastTicketSent  *ticketer.Ticket
}

func (td *testDispatcher) ID() string {
	return td.id
}

func (td *testDispatcher) Roads() []ticketer.RoadID {
	return td.roads
}

func (td *testDispatcher) SendTicket(t *ticketer.Ticket) {
	td.sentTicketCount++
	td.lastTicketSent = t
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
		wantDispatcherCounts map[ticketer.RoadID]int
	}{
		{
			name:                 "Single dispatcher",
			dispatchers:          []ticketer.Dispatcher{&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}}},
			wantDispatcherCounts: map[ticketer.RoadID]int{1: 1, 2: 1, 3: 1},
		},
		{
			name: "Multiple dispatchers",
			dispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}},
				&testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{2}},
				&testDispatcher{id: "dispatcher_3", roads: []ticketer.RoadID{2, 3}},
				&testDispatcher{id: "dispatcher_4", roads: []ticketer.RoadID{10}},
				&testDispatcher{id: "dispatcher_5", roads: []ticketer.RoadID{10, 11, 2}},
			},
			wantDispatcherCounts: map[ticketer.RoadID]int{1: 1, 2: 4, 3: 2, 10: 2, 11: 1},
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
		wantDispatcherCounts map[ticketer.RoadID]int
	}{
		{
			name:                 "Remove dispatcher that does not exist",
			dispatchers:          []ticketer.Dispatcher{&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}}},
			wantDispatcherCounts: map[ticketer.RoadID]int{1: 0, 2: 0, 3: 0},
		},
		{
			name:                 "Add then remove dispatcher",
			startDispatchers:     []ticketer.Dispatcher{&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}}},
			dispatchers:          []ticketer.Dispatcher{&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}}},
			wantDispatcherCounts: map[ticketer.RoadID]int{1: 0, 2: 0, 3: 0},
		},
		{
			name: "Add then remove multiple dispatchers",
			startDispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}},
				&testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{2}},
				&testDispatcher{id: "dispatcher_3", roads: []ticketer.RoadID{2, 3}},
				&testDispatcher{id: "dispatcher_4", roads: []ticketer.RoadID{10}},
				&testDispatcher{id: "dispatcher_5", roads: []ticketer.RoadID{10, 11, 2}},
			},
			dispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}},
				&testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{2}},
				&testDispatcher{id: "dispatcher_3", roads: []ticketer.RoadID{2, 3}},
				&testDispatcher{id: "dispatcher_4", roads: []ticketer.RoadID{10}},
				&testDispatcher{id: "dispatcher_5", roads: []ticketer.RoadID{10, 11, 2}},
			},
			wantDispatcherCounts: map[ticketer.RoadID]int{1: 0, 2: 0, 3: 0, 10: 0, 11: 0},
		},
		{
			name: "Add then remove some dispatchers",
			startDispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}},
				&testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{2}},
				&testDispatcher{id: "dispatcher_3", roads: []ticketer.RoadID{2, 3}},
				&testDispatcher{id: "dispatcher_4", roads: []ticketer.RoadID{10}},
				&testDispatcher{id: "dispatcher_5", roads: []ticketer.RoadID{10, 11, 2}},
			},
			dispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{2}},
				&testDispatcher{id: "dispatcher_3", roads: []ticketer.RoadID{2, 3}},
				&testDispatcher{id: "dispatcher_5", roads: []ticketer.RoadID{10, 11, 2}},
			},
			wantDispatcherCounts: map[ticketer.RoadID]int{1: 1, 2: 1, 3: 1, 10: 1, 11: 0},
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

func TestLocateDispatcher(t *testing.T) {
	tt := []struct {
		name             string
		startDispatchers []ticketer.Dispatcher
		roads            []ticketer.RoadID
		wantDispatchers  []ticketer.Dispatcher
	}{
		{
			name:            "LocateDispatcher returns nil if no dispatcher found for road",
			roads:           []ticketer.RoadID{100},
			wantDispatchers: []ticketer.Dispatcher{nil},
		},
		{
			name:             "Dispatcher located for each road",
			startDispatchers: []ticketer.Dispatcher{&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}}},
			roads:            []ticketer.RoadID{1, 2, 3},
			wantDispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}},
				&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}},
				&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}},
			},
		},
		{
			name: "Dispatcher registered first is returned first every time",
			startDispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{3, 4, 5}},
				&testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2, 3}},
			},
			roads: []ticketer.RoadID{3, 3, 3},
			wantDispatchers: []ticketer.Dispatcher{
				&testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{3, 4, 5}},
				&testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{3, 4, 5}},
				&testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{3, 4, 5}},
			},
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

			for i, r := range tc.roads {
				is.Equal(tm.LocateDispatcher(r), tc.wantDispatchers[i]) // dispatcher mismatch
			}
		})
	}
}

func TestAttemptIdenticalTicketIssueWithNoDispatchers(t *testing.T) {
	is := is.New(t)

	tm := ticketer.NewTicketManager()

	is.Equal(len(tm.SentTickets), 0)   // expected no sent tickets
	is.Equal(len(tm.UnsentTickets), 0) // expected no unsent tickets

	ticket := ticketer.NewTicket("HELLO", ticketer.RoadID(1), 50, 51, 1001800, 1001835, 10300)
	tm.AttemptTicketIssue(ticket)

	is.Equal(len(tm.SentTickets), 0)                       // expected no sent tickets
	is.Equal(len(tm.UnsentTickets), 1)                     // expected unsent tickets for 1 road
	is.Equal(len(tm.UnsentTickets[ticketer.RoadID(1)]), 1) // expected 1 unsent ticket

	ticket = ticketer.NewTicket("HELLO", ticketer.RoadID(1), 50, 51, 1001800, 1001835, 10300)
	tm.AttemptTicketIssue(ticket)

	ticket = ticketer.NewTicket("HELLO", ticketer.RoadID(1), 50, 51, 1001800, 1001835, 10300)
	tm.AttemptTicketIssue(ticket)

	is.Equal(len(tm.SentTickets), 0)                       // expected no sent tickets
	is.Equal(len(tm.UnsentTickets), 1)                     // expected unsent tickets for 1 road
	is.Equal(len(tm.UnsentTickets[ticketer.RoadID(1)]), 3) // expected 3 unsent ticket
}

func TestAttemptTicketIssueWithConnectedDispatcher(t *testing.T) {
	is := is.New(t)

	tm := ticketer.NewTicketManager()

	dispatcher1 := &testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2}}
	tm.AddDispatcher(dispatcher1)

	is.Equal(len(tm.SentTickets), 0)          // expected no sent tickets
	is.Equal(len(tm.UnsentTickets), 0)        // expected no unsent tickets
	is.Equal(dispatcher1.sentTicketCount, 0)  // expected ticket to not be sent yet
	is.Equal(dispatcher1.lastTicketSent, nil) // sent ticket mismatch

	ticket := ticketer.NewTicket("HELLO", ticketer.RoadID(1), 50, 51, 1_001_800, 1_001_835, 10300)
	tm.AttemptTicketIssue(ticket)

	is.Equal(len(tm.SentTickets), 1)             // expected 1 sent ticket
	is.Equal(len(tm.UnsentTickets), 0)           // expected no unsent tickets
	is.Equal(dispatcher1.sentTicketCount, 1)     // expected ticket to be sent
	is.Equal(dispatcher1.lastTicketSent, ticket) // sent ticket mismatch
}

func TestIssueMaxOneTicketPerPlateDayCombination(t *testing.T) {
	is := is.New(t)

	tm := ticketer.NewTicketManager()

	dispatcher1 := &testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2}}
	tm.AddDispatcher(dispatcher1)

	ticket1 := ticketer.NewTicket("HELLO", ticketer.RoadID(1), 50, 51, 1_001_800, 1_001_835, 10300)
	tm.AttemptTicketIssue(ticket1)

	is.Equal(len(tm.SentTickets), 1)              // expected 1 sent ticket
	is.Equal(len(tm.UnsentTickets), 0)            // expected no unsent tickets
	is.Equal(dispatcher1.sentTicketCount, 1)      // expected ticket to be sent
	is.Equal(dispatcher1.lastTicketSent, ticket1) // sent ticket mismatch

	// compared with ticket1, same plate, same day - road is different but doesn't matter
	ticket2 := ticketer.NewTicket("HELLO", ticketer.RoadID(2), 10, 20, 950_400, 951_000, 10300)
	tm.AttemptTicketIssue(ticket2)

	is.Equal(len(tm.SentTickets), 1)              // expected 1 sent ticket
	is.Equal(len(tm.UnsentTickets), 0)            // expected no unsent tickets
	is.Equal(dispatcher1.sentTicketCount, 1)      // expected no change in number of tickets sent
	is.Equal(dispatcher1.lastTicketSent, ticket1) // sent ticket mismatch

	// compared with ticket1, same plate, new day
	ticket3 := ticketer.NewTicket("HELLO", ticketer.RoadID(2), 10, 20, 1_036_800, 1_037_800, 10300)
	tm.AttemptTicketIssue(ticket3)

	is.Equal(len(tm.SentTickets), 2)              // expected 2 sent tickets
	is.Equal(len(tm.UnsentTickets), 0)            // expected no unsent tickets
	is.Equal(dispatcher1.sentTicketCount, 2)      // expected ticket to be sent
	is.Equal(dispatcher1.lastTicketSent, ticket3) // sent ticket mismatch

	// compared with ticket1, new plate, same day
	ticket4 := ticketer.NewTicket("WORLD", ticketer.RoadID(1), 50, 51, 1_001_800, 1_001_835, 10300)
	tm.AttemptTicketIssue(ticket4)

	is.Equal(len(tm.SentTickets), 3)              // expected 3 sent tickets
	is.Equal(len(tm.UnsentTickets), 0)            // expected no unsent tickets
	is.Equal(dispatcher1.sentTicketCount, 3)      // expected ticket to be sent
	is.Equal(dispatcher1.lastTicketSent, ticket4) // sent ticket mismatch
}

func TestAddingDispatcherTriggersTicketIssue(t *testing.T) {
	is := is.New(t)

	tm := ticketer.NewTicketManager()

	ticket1 := ticketer.NewTicket("HELLO", ticketer.RoadID(1), 50, 51, 1_001_800, 1_001_835, 10300)
	ticket2 := ticketer.NewTicket("WORLD", ticketer.RoadID(1), 50, 51, 1_001_800, 1_001_835, 10300)
	ticket3 := ticketer.NewTicket("HELLO", ticketer.RoadID(2), 50, 51, 1_209_600, 1_210_600, 10300)
	ticket4 := ticketer.NewTicket("APPLE", ticketer.RoadID(2), 50, 51, 1_001_800, 1_001_835, 10300)
	ticket5 := ticketer.NewTicket("MANGO", ticketer.RoadID(3), 50, 51, 1_001_800, 1_001_835, 10300)

	tm.AttemptTicketIssue(ticket1)
	tm.AttemptTicketIssue(ticket2)
	tm.AttemptTicketIssue(ticket3)
	tm.AttemptTicketIssue(ticket4)
	tm.AttemptTicketIssue(ticket5)

	is.Equal(len(tm.SentTickets), 0)                       // expected no sent tickets
	is.Equal(len(tm.UnsentTickets), 3)                     // expected unsent tickets for 3 roads
	is.Equal(len(tm.UnsentTickets[ticketer.RoadID(1)]), 2) // expected 2 unsent tickets for road 1
	is.Equal(len(tm.UnsentTickets[ticketer.RoadID(2)]), 2) // expected 2 unsent tickets for road 2
	is.Equal(len(tm.UnsentTickets[ticketer.RoadID(3)]), 1) // expected 1 unsent ticket for road 3

	dispatcher1 := &testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1, 2}}
	tm.AddDispatcher(dispatcher1)

	is.Equal(len(tm.SentTickets), 4)         // expected 4 sent tickets
	is.Equal(len(tm.UnsentTickets), 1)       // expected unsent tickets for 1 road
	is.Equal(dispatcher1.sentTicketCount, 4) // expected 4 tickets to be received by dispatcher1

	dispatcher2 := &testDispatcher{id: "dispatcher_2", roads: []ticketer.RoadID{1, 2, 3}}
	tm.AddDispatcher(dispatcher2)

	is.Equal(len(tm.SentTickets), 5)         // expected 5 sent tickets
	is.Equal(len(tm.UnsentTickets), 0)       // expected 0 unsent tickets
	is.Equal(dispatcher1.sentTicketCount, 4) // expected dispatcher1 to still be at 4 tickets
	is.Equal(dispatcher2.sentTicketCount, 1) // expected 1 ticket to be received by dispatcher2
}

func TestObserveWillAttemptToIssuesTickets(t *testing.T) {
	is := is.New(t)

	tm := ticketer.NewTicketManager()

	dispatcher1 := &testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{1}}
	tm.AddDispatcher(dispatcher1)

	is.Equal(len(tm.SentTickets), 0)   // expected no sent tickets
	is.Equal(len(tm.UnsentTickets), 0) // expected no unsent tickets

	road1 := &ticketer.Road{ID: ticketer.RoadID(1), Limit: 50}

	observation1Road1 := &ticketer.Observation{
		Road:      road1,
		Mile:      0,
		Plate:     "ABC123",
		Timestamp: 0,
	}
	tm.Observe(observation1Road1)

	// no ticket after single observation
	is.Equal(len(tm.SentTickets), 0)   // expected no sent tickets
	is.Equal(len(tm.UnsentTickets), 0) // expected no unsent tickets

	observation2Road1 := &ticketer.Observation{
		Road:      road1,
		Mile:      10,
		Plate:     "ABC123",
		Timestamp: 700, // 720 is on speed limit
	}
	tm.Observe(observation2Road1)

	is.Equal(len(tm.SentTickets), 1)         // expected 1 sent ticket
	is.Equal(len(tm.UnsentTickets), 0)       // expected no unsent tickets
	is.Equal(dispatcher1.sentTicketCount, 1) // expected 1 ticket to be received by dispatcher1

	// overlapping observation
	observation3Road1 := &ticketer.Observation{
		Road:      road1,
		Mile:      9,
		Plate:     "ABC123",
		Timestamp: 300, // 648 is on speed limit
	}
	tm.Observe(observation3Road1)

	is.Equal(len(tm.SentTickets), 1)         // expected no more sent tickets
	is.Equal(len(tm.UnsentTickets), 0)       // expected no unsent tickets
	is.Equal(dispatcher1.sentTicketCount, 1) // expected dispatcher1 to still be at 1 ticket

	road2 := &ticketer.Road{ID: ticketer.RoadID(2), Limit: 10}

	observation1Road2 := &ticketer.Observation{
		Road:      road2,
		Mile:      0,
		Plate:     "DEF456",
		Timestamp: 0,
	}

	observation2Road2 := &ticketer.Observation{
		Road:      road2,
		Mile:      100,
		Plate:     "DEF456",
		Timestamp: 2800,
	}

	tm.Observe(observation1Road2)
	tm.Observe(observation2Road2)

	is.Equal(len(tm.SentTickets), 1)   // expected no more sent tickets
	is.Equal(len(tm.UnsentTickets), 1) // expected 1 unsent ticket
}

func TestMultiDayScenario(t *testing.T) {
	is := is.New(t)

	tm := ticketer.NewTicketManager()

	road := &ticketer.Road{ID: ticketer.RoadID(18452), Limit: 50}
	plate := "CT77YVD"

	dispatcher1 := &testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{road.ID}}
	tm.AddDispatcher(dispatcher1)

	// Car speeding consistently at ~62 mph. Every combination of observations is a possible
	// ticket. The days over which a ticket spans cannot overlap with the span of another
	// ticket. First ticket will span Day 812 and 813 and should issue after the first pair
	// of observations. No other tickets should be issued.
	observations := []*ticketer.Observation{
		{Road: road, Mile: 351, Plate: plate, Timestamp: 70233988}, // Day 812
		{Road: road, Mile: 650, Plate: plate, Timestamp: 70251461}, // Day 813
		{Road: road, Mile: 460, Plate: plate, Timestamp: 70240358}, // Day 812
		{Road: road, Mile: 905, Plate: plate, Timestamp: 70266363}, // Day 813
	}

	for _, o := range observations {
		tm.Observe(o)
	}

	is.Equal(dispatcher1.sentTicketCount, 1) // expected 1 sent ticket
}

func TestAnotherTicketScenario(t *testing.T) {
	is := is.New(t)

	tm := ticketer.NewTicketManager()

	road := &ticketer.Road{ID: ticketer.RoadID(25958), Limit: 45}
	plate := "NN90MLU"

	dispatcher1 := &testDispatcher{id: "dispatcher_1", roads: []ticketer.RoadID{road.ID}}
	tm.AddDispatcher(dispatcher1)

	observations := []*ticketer.Observation{
		{Road: road, Mile: 183, Plate: plate, Timestamp: 45343155}, // Day 524
		{Road: road, Mile: 423, Plate: plate, Timestamp: 45330449}, // Day 524
		{Road: road, Mile: 345, Plate: plate, Timestamp: 45361060}, // Day 525
		{Road: road, Mile: 631, Plate: plate, Timestamp: 45371356}, // Day 525
		{Road: road, Mile: 130, Plate: plate, Timestamp: 45353320}, // Day 524
		{Road: road, Mile: 938, Plate: plate, Timestamp: 45382408}, // Day 525
	}

	for _, o := range observations {
		tm.Observe(o)
	}

	is.Equal(dispatcher1.sentTicketCount, 2) // expected 2 sent tickets
}
