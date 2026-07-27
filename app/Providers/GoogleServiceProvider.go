package Providers

import (
	"github.com/rancago/framework/app/Contracts"
	"github.com/rancago/framework/framework/Container"
	"github.com/rancago/framework/framework/Google"
)

// GoogleServiceProvider registers Calendar, Meet, and MeetingScheduler services.
//
// Container bindings (post-registration):
//   - "google.calendar"              → Contracts.CalendarService (singleton)
//   - "google.meet"                  → Contracts.MeetService (singleton)
//   - "google.scheduler"             → Contracts.MeetingScheduler (singleton)
//   - "Contracts.CalendarService"    → alias
//   - "Contracts.MeetService"        → alias
//   - "Contracts.MeetingScheduler"   → alias
type GoogleServiceProvider struct {
	cfg Google.GoogleConfig
}

// NewGoogleServiceProvider creates a GoogleServiceProvider.
func NewGoogleServiceProvider(cfg Google.GoogleConfig) *GoogleServiceProvider {
	return &GoogleServiceProvider{cfg: cfg}
}

// Register binds Google services into the container.
func (p *GoogleServiceProvider) Register(c *Container.Container) error {
	c.Singleton("google.calendar", func(_ *Container.Container) (interface{}, error) {
		return Google.NewCalendarService(&p.cfg), nil
	})
	c.Alias("google.calendar", "Contracts.CalendarService")

	c.Singleton("google.meet", func(_ *Container.Container) (interface{}, error) {
		return Google.NewMeetService(&p.cfg), nil
	})
	c.Alias("google.meet", "Contracts.MeetService")

	c.Singleton("google.scheduler", func(c *Container.Container) (interface{}, error) {
		calRaw, err := c.Resolve("google.calendar")
		if err != nil {
			return nil, err
		}
		meetRaw, err := c.Resolve("google.meet")
		if err != nil {
			return nil, err
		}
		return Google.NewMeetingScheduler(
			calRaw.(Contracts.CalendarService),
			meetRaw.(Contracts.MeetService),
		), nil
	})
	c.Alias("google.scheduler", "Contracts.MeetingScheduler")
	return nil
}

// Boot is a no-op — all wiring happens in Register.
func (p *GoogleServiceProvider) Boot(_ *Container.Container) error { return nil }
