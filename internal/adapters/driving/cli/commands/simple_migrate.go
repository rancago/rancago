package commands

import (
	"flag"
	"fmt"
)

type SimpleCommand struct {
	use     string
	short   string
	argsMin int
	argsMax int
	flags   *flag.FlagSet
	runFn   func(fs *flag.FlagSet, args []string) error
}

func (c *SimpleCommand) SetArgs(a []string) {
	_ = c.flags.Parse(a)
}

func (c *SimpleCommand) Execute() error {
	args := c.flags.Args()
	if c.argsMin > 0 && len(args) < c.argsMin {
		return fmt.Errorf("usage: gawego %s — %s (missing required args)", c.use, c.short)
	}
	if c.argsMax >= 0 && len(args) > c.argsMax {
		return fmt.Errorf("usage: gawego %s — too many arguments", c.use)
	}
	return c.runFn(c.flags, args)
}

func NewMigrateCommand() *SimpleCommand {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	rollback := fs.Bool("rollback", false, "Rollback the last batch of migrations")
	return &SimpleCommand{
		use:     "migrate",
		short:   "Run pending database migrations",
		argsMin: 0,
		argsMax: 0,
		flags:   fs,
		runFn: func(f *flag.FlagSet, _ []string) error {
			action := "applying"
			if *rollback {
				action = "rolling back"
			}
			fmt.Printf("\n  🗄️  Migrate: %s migrations\n", action)
			fmt.Println("  " + stringsRepeat("=", 60))
			fmt.Println("  Using in-memory persistence driver — plug in a real DB driven adapter (GORM/sql.DB) for migrations")
			fmt.Println("  Migration driver: mock (skipped)")
			fmt.Println("  Migration status: OK (no-op)")
			fmt.Println()
			return nil
		},
	}
}
