package ticketer

import (
	"math"
)

type TicketManager struct {
	Observations  map[observationKey][]*Observation
	Dispatchers   map[RoadID][]Dispatcher
	SentTickets   map[ticketKey]*Ticket
	UnsentTickets map[RoadID][]*Ticket
}

func NewTicketManager() *TicketManager {
	return &TicketManager{
		Observations:  map[observationKey][]*Observation{},
		Dispatchers:   map[RoadID][]Dispatcher{},
		SentTickets:   map[ticketKey]*Ticket{},
		UnsentTickets: map[RoadID][]*Ticket{},
	}
}

func (t *TicketManager) comparable(o1 *Observation, o2 *Observation) bool {
	return o1 != nil && o2 != nil && o1 != o2 && o1.comparable(o2)
}

func (t *TicketManager) Observe(o *Observation) bool {
	if o == nil {
		return false
	}

	key := o.key()

	isMatch := false
	if obs, ok := t.Observations[key]; ok {
		isMatch = len(obs) > 0
	}

	t.Observations[key] = append(t.Observations[key], o)

	return isMatch
}

func (t *TicketManager) DetectSpeedingInfractions(o *Observation) []*Ticket {
	tickets := []*Ticket{}

	if o != nil {
		if obs, ok := t.Observations[o.key()]; ok {
			for _, oOther := range obs {
				speed, isSpeeding := t.DetectSpeeding(o, oOther)
				if isSpeeding {
					// order by mile (ascending)
					o1, o2 := o, oOther
					if o.Timestamp > oOther.Timestamp {
						o1, o2 = oOther, o
					}
					ticket := NewTicket(o1.Plate, o1.Road.ID, o1.Mile, o2.Mile, o1.Timestamp, o2.Timestamp, speed)
					tickets = append(tickets, ticket)
				}
			}
		}
	}

	return tickets
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

func (t *TicketManager) DetectSpeeding(o1 *Observation, o2 *Observation) (uint16, bool) {
	speed := t.CalculateSpeed(o1, o2)
	return speed, t.comparable(o1, o2) && speed > o1.Road.Limit
}

func (t *TicketManager) AttemptTicketIssue(ticket *Ticket) {
	// has ticket been issued today for plate?
	if _, ok := t.SentTickets[ticket.key()]; !ok {
		if dispatcher := t.LocateDispatcher(ticket.Road); dispatcher != nil {
			dispatcher.SendTicket(ticket)
			t.SentTickets[ticket.key()] = ticket
		} else {
			t.UnsentTickets[ticket.Road] = append(t.UnsentTickets[ticket.Road], ticket)
		}
	}
}

func (t *TicketManager) HasUnsentTicket(r *Road) *Ticket {
	// TODO
	return nil
}

func (t *TicketManager) AddDispatcher(d Dispatcher) {
	// associate each Dispatcher with every road its responsible for
	for _, r := range d.Roads() {
		found := false
		for _, dispatcher := range t.Dispatchers[r] {
			if dispatcher.ID() == d.ID() {
				found = true
				break
			}
		}
		if !found {
			t.Dispatchers[r] = append(t.Dispatchers[r], d)
			// TODO: any unsent tickets for this road?
		}
	}
}

func (t *TicketManager) RemoveDispatcher(d Dispatcher) {
	for _, r := range d.Roads() {
		rds := t.Dispatchers[r]
		for i, dispatcher := range rds {
			if dispatcher.ID() == d.ID() {
				rds[i] = rds[len(rds)-1]
				t.Dispatchers[r] = rds[:len(rds)-1]
				break
			}
		}
	}
}

func (t *TicketManager) LocateDispatcher(roadID RoadID) Dispatcher {
	// return first dispatcher found for road
	if dispatchers, ok := t.Dispatchers[roadID]; ok {
		for _, d := range dispatchers {
			return d
		}
	}
	return nil
}
