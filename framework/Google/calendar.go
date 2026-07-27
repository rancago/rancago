// Package Google provides Google Calendar and Google Meet integrations for Rancago.
// Currently ships a stub implementation - swap in google.golang.org/api once added to go.mod.
package Google

import (
	"context"
	"fmt"
	"time"

	"github.com/rancago/framework/app/Contracts"
)

// GoogleConfig holds Google API credentials.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Credentials  string // path to service account JSON
}

// CalendarService is the Google Calendar adapter.
// It satisfies Contracts.CalendarService.
type CalendarService struct {
	cfg *GoogleConfig
}

// NewCalendarService creates a CalendarService.
func NewCalendarService(cfg *GoogleConfig) Contracts.CalendarService {
	return &CalendarService{cfg: cfg}
}

func (s *CalendarService) CreateEvent(_ context.Context, ev *Contracts.CalendarEvent) (*Contracts.CalendarEvent, error) {
	ev.ID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
	ev.HTMLLink = fmt.Sprintf("https://calendar.google.com/event?eid=%s", ev.ID)
	if ev.ConferenceData != nil && ev.ConferenceData.CreateLink {
		ev.MeetLink = fmt.Sprintf("https://meet.google.com/%s", randomCode())
	}
	return ev, nil
}

func (s *CalendarService) UpdateEvent(_ context.Context, eventID string, ev *Contracts.CalendarEvent) (*Contracts.CalendarEvent, error) {
	ev.ID = eventID
	return ev, nil
}

func (s *CalendarService) DeleteEvent(_ context.Context, _ string) error { return nil }

func (s *CalendarService) GetEvent(_ context.Context, eventID string) (*Contracts.CalendarEvent, error) {
	return &Contracts.CalendarEvent{
		ID:      eventID,
		Summary: "Rancago Event (stub)",
		Start:   time.Now(),
		End:     time.Now().Add(time.Hour),
	}, nil
}

func (s *CalendarService) ListEvents(_ context.Context, from, to time.Time) ([]*Contracts.CalendarEvent, error) {
	return []*Contracts.CalendarEvent{
		{
			ID:       "sample_event_1",
			Summary:  "Rancago Demo Sync",
			Start:    from,
			End:      from.Add(time.Hour),
			MeetLink: "https://meet.google.com/abc-defg-hij",
		},
	}, nil
}

func (s *CalendarService) AddAttendees(_ context.Context, eventID string, attendees []Contracts.Attendee) (*Contracts.CalendarEvent, error) {
	return &Contracts.CalendarEvent{
		ID:        eventID,
		Attendees: attendees,
	}, nil
}
