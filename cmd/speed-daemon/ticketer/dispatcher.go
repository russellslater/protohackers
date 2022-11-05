package ticketer

type Dispatcher interface {
	ID() string
	Roads() []RoadID
	SendTicket(t *Ticket)
}
