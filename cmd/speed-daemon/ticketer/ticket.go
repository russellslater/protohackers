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

func (t *Ticket) String() string {
	return fmt.Sprintf("Ticket for %s (road=%d, mm1=%d, ts1=%d, mm2=%d, ts2=%d, speed=%d]\n",
		t.Plate, t.Road, t.MileStart, t.TimestampStart, t.MileEnd, t.TimestampEnd, t.Speed/100)
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

func (t *Ticket) SpannedDays() []int {
	spannedDays := []int{}
	for i := t.StartDay(); i <= t.EndDay(); i++ {
		spannedDays = append(spannedDays, i)
	}
	return spannedDays
}
