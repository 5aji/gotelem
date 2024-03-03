package db
// This file implements Packet modelling, which allows us to look up fields by name

type PacketDef struct {
	Name string
	Description string
	Id int
}

type FieldDef struct {
	Name string
	SubName string
	Packet string
	Type string
}

// PacketNotFoundError is when a matching packet cannot be found.
type PacketNotFoundError string

func (e *PacketNotFoundError) Error() string {
	return "packet not found: " + string(*e)
}


// GetPacketDefN retrieves a packet matching the given name, if it exists.
// returns PacketNotFoundError if a matching packet could not be found.
func (tdb *TelemDb) GetPacketDefN(name string) (*PacketDef, error) {
	return nil, nil
}

// GetPacketDefF retrieves the parent packet for a given field. 
// This function cannot return PacketNotFoundError since we have SQL FKs enforcing.
func (tdb *TelemDb) GetPacketDefF(field FieldDef) (*PacketDef, error) {
	return nil, nil 
}


// GetFieldDefs returns the given fields for a given packet definition.
func (tdb *TelemDb) GetFieldDefs(pkt PacketDef) ([]FieldDef, error) {
	return nil, nil
}


