package google

import (
	"context"
	"fmt"
	"time"

	"github.com/rancago/framework/internal/kernel"
)

type CalendarEvent struct {
	ID          string
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
	Attendees   []string
	MeetLink    string
}

type CalendarAdapter struct {
	cfg *kernel.GoogleConfig
}

func NewCalendarAdapter(cfg *kernel.GoogleConfig) *CalendarAdapter {
	return &CalendarAdapter{cfg: cfg}
}

func (a *CalendarAdapter) CreateEvent(_ context.Context, ev CalendarEvent) (*CalendarEvent, error) {
	ev.ID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
	ev.MeetLink = fmt.Sprintf("https://meet.google.com/%s", randomMeetCode())
	return &ev, nil
}

func (a *CalendarAdapter) ListEvents(_ context.Context, from, to time.Time) ([]CalendarEvent, error) {
	return []CalendarEvent{
		{
			ID:        "sample_1",
			Title:     "Gawego Demo Sync",
			StartTime: from,
			EndTime:   from.Add(1 * time.Hour),
			MeetLink:  "https://meet.google.com/abc-defg-hij",
		},
	}, nil
}

type MeetSpace struct {
	ID         string
	URI        string
	MeetingCode string
	CreatorID  string
}

type MeetAdapter struct {
	cfg *kernel.GoogleConfig
}

func NewMeetAdapter(cfg *kernel.GoogleConfig) *MeetAdapter {
	return &MeetAdapter{cfg: cfg}
}

func (a *MeetAdapter) CreateSpace(_ context.Context, creatorID string, name string) (*MeetSpace, error) {
	code := randomMeetCode()
	return &MeetSpace{
		ID:         "space_" + code,
		URI:        fmt.Sprintf("https://meet.google.com/%s", code),
		MeetingCode: code,
		CreatorID:  creatorID,
	}, nil
}

func (a *MeetAdapter) EndSpace(_ context.Context, spaceID string) error {
	return nil
}

func randomMeetCode() string {
	now := time.Now().UnixNano()
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 10)
	for i := 0; i < 10; i++ {
		result[i] = alphabet[int(now>>uint(i*3))%len(alphabet)]
	}
	return fmt.Sprintf("%s-%s-%s", string(result[:3]), string(result[3:7]), string(result[7:]))
}
