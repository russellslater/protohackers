package ticketer

import "math"

type Road struct {
	ID    uint16
	Limit uint16
}

type Observation struct {
	Road      *Road
	Mile      uint16
	Plate     string
	Timestamp uint32
}

func (o *Observation) comparable(b *Observation) bool {
	return (o.Road.ID == b.Road.ID) && (o.Plate == b.Plate)
}

type observationKey struct {
	roadID uint16
	plate  string
}

type Dispatcher interface {
	Road() *Road
	SendTicket(t *Ticket)
}

type Ticket struct {
	Plate          string
	Road           uint16
	MileStart      uint16
	MileEnd        uint16
	TimestampStart uint32
	TimestampEnd   uint32
	Speed          uint16 // 100x mile per hour
}

type TicketManager struct {
	Observations  map[observationKey][]*Observation
	Dispatcher    []*Dispatcher
	UnsentTickets []*Ticket
}

func NewTicketManager() *TicketManager {
	return &TicketManager{
		Observations: map[observationKey][]*Observation{},
	}
}

func (t *TicketManager) comparable(o1 *Observation, o2 *Observation) bool {
	return o1 != nil && o2 != nil && o1 != o2 && o1.comparable(o2)
}

func (t *TicketManager) Observe(o *Observation) bool {
	if o == nil {
		return false
	}

	key := observationKey{roadID: o.Road.ID, plate: o.Plate}

	isMatch := false
	if obs, ok := t.Observations[key]; ok {
		isMatch = len(obs) > 0
	}

	t.Observations[key] = append(t.Observations[key], o)

	return isMatch
}

func (t *TicketManager) CalculateSpeed(o1 *Observation, o2 *Observation) uint16 {
	speed := 0

	if t.comparable(o1, o2) {
		distance := math.Abs(float64(o1.Mile) - float64(o2.Mile))
		timeSeconds := math.Abs(float64(o1.Timestamp) - float64(o2.Timestamp))

		// speed = distance / time
		speed = int(math.Round(distance / (timeSeconds / 3600))) // miles per hour
	}

	return uint16(speed)
}

func (t *TicketManager) DetectSpeeding(o1 *Observation, o2 *Observation) bool {
	return t.comparable(o1, o2) && t.CalculateSpeed(o1, o2) > o1.Road.Limit
}

func (t *TicketManager) WriteTicket(o1 *Observation, o2 *Observation) *Ticket {
	// TODO
	// day boudaries - floor(timestamp / 86400)
	return nil
}

func (t *TicketManager) HasUnsentTicket(r *Road) *Ticket {
	// TODO
	return nil
}

func (t *TicketManager) AddDispatcher(d *Dispatcher) {
	// TODO
}

func (t *TicketManager) LocateDispatcher(r *Road) *Dispatcher {
	// TODO
	return nil
}
