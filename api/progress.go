package api

// ProgressFunc is used to output events
type ProgressFunc func(event *Event, err error)

// Event contains the properties of a server sent event
type Event struct {
	ID          string `json:"id"`
	MessageType string
	Message     string `json:"message"`
	Completed   int    `json:"completed"`
	Total       int    `json:"total"`
}
