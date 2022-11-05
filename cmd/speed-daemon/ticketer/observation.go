package ticketer

type Observation struct {
	Road      *Road
	Mile      uint16
	Plate     string
	Timestamp uint32
}

func (o *Observation) comparable(b *Observation) bool {
	return o.key() == b.key()
}

func (o *Observation) key() observationKey {
	return observationKey{roadID: o.Road.ID, plate: o.Plate}
}

type observationKey struct {
	roadID RoadID
	plate  string
}
