package model

import "testing"

func TestGetDefaultFlightList(t *testing.T) {
	list := GetDefaultFlightList()
	if len(list) == 0 {
		t.Fail()
	}
}

func TestFlight_AddBooking(t *testing.T) {
	bLevelLow := NewBonusLevel(utLowLevelID, 4, 4)
	bLevelHigh := NewBonusLevel("high", 4, 4)
	flight := &Flight{ID: 1, Bookings: []*Booking{}}
	flight.AddBooking(2, bLevelLow)
	flight.AddBooking(2, bLevelHigh)
	if len(flight.Bookings) != 2 {
		t.Error("Wrong number of flights appended")
		t.Fail()
	}
	if flight.Bookings[0].BonusLevel != bLevelLow {
		t.Error("wrong bonus level added")
		t.Fail()
	}
	if flight.Bookings[1].BonusLevel != bLevelHigh {
		t.Error("wrong bonus level 'high' added")
		t.Fail()
	}
	s := NewServer()
	if _, err := s.Booking(1, 2, "low"); err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(s.flightMap[1].Bookings) != 1 {
		t.Error("wrong length of flight map")
		t.Fail()
	}
}
