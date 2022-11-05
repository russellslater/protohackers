package ticketer

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
	Observations  []*Observation
	Dispatcher    []*Dispatcher
	UnsentTickets []*Ticket
}

func NewTicketManager() *TicketManager {
	return &TicketManager{
		Observations: []*Observation{},
	}
}

func (t *TicketManager) Observe(o *Observation) {
	// TODO
}

func (t *TicketManager) Detect(o1 *Observation, o2 *Observation) bool {
	// TODO
	return true
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
