package ticketer

import (
	"fmt"
	"math"
)

type Ticket struct {
	Plate          string
	Road           RoadID
	MileStart      uint16
	MileEnd        uint16
	TimestampStart uint32
	TimestampEnd   uint32
	Speed          uint16 // 100x mile per hour
}

type ticketKey struct {
	plate string
	day   int
}

func (t *Ticket) key() ticketKey {
	return ticketKey{plate: t.Plate, day: t.StartDay()}
}

func (t *Ticket) String() string {
	return fmt.Sprintf("{P: %s, R: %d, M0: %d, M1: %d, M0: %d, M1: %d, S: %d}",
		t.Plate, t.Road, t.MileStart, t.MileEnd, t.TimestampStart, t.TimestampEnd, t.Speed)
}

func NewTicket(plate string, road RoadID, mileStart uint16, mileEnd uint16, timestampStart uint32, timestampEnd uint32, speed uint16) *Ticket {
	return &Ticket{
		Plate:          plate,
		Road:           road,
		MileStart:      mileStart,
		MileEnd:        mileEnd,
		TimestampStart: timestampStart,
		TimestampEnd:   timestampEnd,
		Speed:          speed * 100,
	}
}

func (t *Ticket) StartDay() int {
	return int(math.Floor(float64(t.TimestampStart) / float64(86400)))
}

func (t *Ticket) EndDay() int {
	return int(math.Floor(float64(t.TimestampEnd) / float64(86400)))
}

func (t *Ticket) DayDiff() int {
	return t.StartDay() - t.EndDay()
}
