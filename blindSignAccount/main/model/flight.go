package model

import "sync"

type Booking struct {
	ID         int
	CustomerID int
	BonusLevel *BonusLevel
}

type Flight struct {
	ID       int
	Bookings []*Booking
	// sync
	mux sync.Mutex
}

// Generate a list of 100 default flights
func GetDefaultFlightList() map[int]*Flight {
	flightList := map[int]*Flight{}
	for i := 0; i < 100; i++ {
		flightList[i] = &Flight{ID: i, Bookings: []*Booking{}}
	}
	return flightList
}

// Adds a new booking to a given flight
func (flight *Flight) AddBooking(customerID int, bLevel *BonusLevel) {
	// sync
	flight.mux.Lock()
	defer flight.mux.Unlock()

	id := len(flight.Bookings) + 1
	booking := &Booking{ID: id, CustomerID: customerID, BonusLevel: bLevel}
	flight.Bookings = append(flight.Bookings, booking)
}
