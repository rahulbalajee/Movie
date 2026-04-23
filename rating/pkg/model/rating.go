package model

// RecordID defines a record id. Together with RecordType identifies unique records across all types
type RecordID string

// RecordType defines a record type. Together with RecordID identifies unique records across all types
type RecordType string

// Existing record types
const (
	RecordTypeMovie = RecordType("movie")
)

// UserID defines a user id
type UserID string

// RatingValue defined value of Rating record
type RatingValue int

// Rating defines an individual rating created by a user for some record
type Rating struct {
	RecordID   RecordID    `json:"recordId"`
	RecordType RecordType  `json:"recordType"`
	UserID     UserID      `json:"userId"`
	Value      RatingValue `json:"value"`
}

type RatingEventType string

type RatingEvent struct {
	Rating
	ProviderId string          `json:"providerId"`
	EventType  RatingEventType `json:"eventType"`
}

const (
	RatingEventTypePut    = RatingEventType("put")
	RatingEventTypeDelete = RatingEventType("delete")
	RatingEventTypeGet    = RatingEventType("get")
)
