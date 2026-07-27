package Google

import (
	"context"
	"fmt"
	"time"

	"github.com/rancago/framework/app/Contracts"
)

// MeetService is the Google Meet adapter.
// It satisfies Contracts.MeetService.
type MeetService struct {
	cfg *GoogleConfig
}

// NewMeetService creates a MeetService.
func NewMeetService(cfg *GoogleConfig) Contracts.MeetService {
	return &MeetService{cfg: cfg}
}

func (s *MeetService) CreateSpace(_ context.Context, opts *Contracts.MeetSpaceOptions) (*Contracts.MeetSpace, error) {
	code := randomCode()
	space := &Contracts.MeetSpace{
		ID:          "space_" + code,
		URI:         fmt.Sprintf("https://meet.google.com/%s", code),
		JoinURL:     fmt.Sprintf("https://meet.google.com/%s", code),
		MeetingCode: code,
	}
	if opts != nil && len(opts.Moderators) > 0 {
		space.CreatorID = opts.Moderators[0]
	}
	return space, nil
}

func (s *MeetService) GetSpace(_ context.Context, spaceID string) (*Contracts.MeetSpace, error) {
	return &Contracts.MeetSpace{
		ID:      spaceID,
		URI:     fmt.Sprintf("https://meet.google.com/%s", spaceID),
		JoinURL: fmt.Sprintf("https://meet.google.com/%s", spaceID),
	}, nil
}

func (s *MeetService) GenerateJoinURL(_ context.Context, spaceID string) (string, error) {
	return fmt.Sprintf("https://meet.google.com/%s", spaceID), nil
}

// MeetingScheduler is the facade that creates a Calendar event + Meet link in one call.
// It satisfies Contracts.MeetingScheduler.
type MeetingSchedulerService struct {
	calendar Contracts.CalendarService
	meet     Contracts.MeetService
}

// NewMeetingScheduler creates a MeetingScheduler backed by the provided services.
func NewMeetingScheduler(calendar Contracts.CalendarService, meet Contracts.MeetService) Contracts.MeetingScheduler {
	return &MeetingSchedulerService{calendar: calendar, meet: meet}
}

func (s *MeetingSchedulerService) ScheduleWithMeet(ctx context.Context, ev *Contracts.CalendarEvent) (*Contracts.MeetingResult, error) {
	created, err := s.calendar.CreateEvent(ctx, ev)
	if err != nil {
		return nil, fmt.Errorf("scheduler: create event: %w", err)
	}
	var space *Contracts.MeetSpace
	if ev.ConferenceData != nil && ev.ConferenceData.CreateLink {
		space, err = s.meet.CreateSpace(ctx, &Contracts.MeetSpaceOptions{Name: ev.Summary})
		if err != nil {
			return nil, fmt.Errorf("scheduler: create meet space: %w", err)
		}
		created.MeetLink = space.JoinURL
	}
	return &Contracts.MeetingResult{CalendarEvent: created, MeetSpace: space}, nil
}

func (s *MeetingSchedulerService) RescheduleWithMeet(ctx context.Context, eventID string, ev *Contracts.CalendarEvent) (*Contracts.MeetingResult, error) {
	updated, err := s.calendar.UpdateEvent(ctx, eventID, ev)
	if err != nil {
		return nil, fmt.Errorf("scheduler: update event: %w", err)
	}
	return &Contracts.MeetingResult{CalendarEvent: updated}, nil
}

func (s *MeetingSchedulerService) CancelMeeting(ctx context.Context, eventID, spaceID string) error {
	return s.calendar.DeleteEvent(ctx, eventID)
}

func randomCode() string {
	now := time.Now().UnixNano()
	alpha := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 10)
	for i := range b {
		b[i] = alpha[int(now>>uint(i*3))%len(alpha)]
	}
	return fmt.Sprintf("%s-%s-%s", string(b[:3]), string(b[3:7]), string(b[7:]))
}
