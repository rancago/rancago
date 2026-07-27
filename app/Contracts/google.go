package Contracts

import (
	"context"
	"time"
)

// Attendee represents a calendar event participant.
type Attendee struct {
	Email          string
	DisplayName    string
	Optional       bool
	ResponseStatus string // accepted, declined, tentative, needsAction
}

// Reminder represents a calendar event reminder.
type Reminder struct {
	Method  string // email | popup
	Minutes int
}

// ConferenceRequest holds the conference creation config for a calendar event.
type ConferenceRequest struct {
	// Type is the conference solution type. Use "hangoutsMeet" for Google Meet.
	Type string
	// CreateLink instructs Google Calendar to auto-create a Meet link.
	CreateLink bool
}

// CalendarEvent is the full representation of a Google Calendar event.
type CalendarEvent struct {
	ID             string
	Summary        string
	Description    string
	Location       string
	Start          time.Time
	End            time.Time
	Timezone       string
	Attendees      []Attendee
	Reminders      []Reminder
	ConferenceData *ConferenceRequest
	Recurrence     []string
	HTMLLink       string
	MeetLink       string
	ICalUID        string
}

// MeetSpaceOptions holds options for creating a standalone Google Meet space.
type MeetSpaceOptions struct {
	Name       string
	Moderators []string
	ExpiresAt  *time.Time
}

// MeetSpace is the result of creating or retrieving a Google Meet space.
type MeetSpace struct {
	ID          string
	URI         string
	JoinURL     string
	MeetingCode string
	CreatorID   string
}

// MeetingResult is the combined result of ScheduleWithMeet.
type MeetingResult struct {
	CalendarEvent *CalendarEvent
	MeetSpace     *MeetSpace
}

// CalendarService manages Google Calendar events.
type CalendarService interface {
	CreateEvent(ctx context.Context, event *CalendarEvent) (*CalendarEvent, error)
	UpdateEvent(ctx context.Context, eventID string, event *CalendarEvent) (*CalendarEvent, error)
	DeleteEvent(ctx context.Context, eventID string) error
	GetEvent(ctx context.Context, eventID string) (*CalendarEvent, error)
	ListEvents(ctx context.Context, from, to time.Time) ([]*CalendarEvent, error)
	AddAttendees(ctx context.Context, eventID string, attendees []Attendee) (*CalendarEvent, error)
}

// MeetService manages standalone Google Meet spaces.
type MeetService interface {
	CreateSpace(ctx context.Context, opts *MeetSpaceOptions) (*MeetSpace, error)
	GetSpace(ctx context.Context, spaceID string) (*MeetSpace, error)
	GenerateJoinURL(ctx context.Context, spaceID string) (string, error)
}

// MeetingScheduler is the high-level facade: 1 call = Calendar event + Meet link.
type MeetingScheduler interface {
	// ScheduleWithMeet creates a calendar event and auto-creates a Meet link when
	// ConferenceData.CreateLink is true.
	ScheduleWithMeet(ctx context.Context, event *CalendarEvent) (*MeetingResult, error)
	// RescheduleWithMeet updates an existing event and refreshes its Meet link.
	RescheduleWithMeet(ctx context.Context, eventID string, event *CalendarEvent) (*MeetingResult, error)
	// CancelMeeting cancels the calendar event and ends the Meet space.
	CancelMeeting(ctx context.Context, eventID, spaceID string) error
}
