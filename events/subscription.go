package events

// Subscription is a read-only event stream.
type Subscription struct {
	id     int
	value  chan any
	events []string
}

// Value returns the next event value from the subscription.
func (s *Subscription) Value() <-chan any {
	return s.value
}

// Events returns the names of all subscribed events.
func (s *Subscription) Events() []string {
	return s.events
}
