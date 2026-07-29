package commands

import (
	"bufio"
	"context"
	"fmt"
	"flag"
	"os"
	"strings"
	"time"
)

func stringsRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func promptLine(r *bufio.Reader, q string) string {
	fmt.Print(q)
	ans, _ := r.ReadString('\n')
	return strings.TrimSpace(ans)
}

func promptBool(r *bufio.Reader, q string, def bool) bool {
	ans := strings.ToLower(promptLine(r, q))
	switch ans {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	}
	return def
}

func NewMakeEntityCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:entity", flag.ContinueOnError)
	return &SimpleCommand{
		use:     "make:entity [name]",
		short:   "Create a new domain entity in internal/domain/entities",
		argsMin: 1,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			gen := Generator{Name: name, Type: "Domain Entity", BasePath: "internal/domain/entities", Package: "entities"}
			mod := ModuleName()
			content := `package entities

import (
	"time"

	"` + mod + `/internal/domain/valueobjects"
)

type ` + pascal + ` struct {
	ID        valueobjects.ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New` + pascal + `() *` + pascal + ` {
	now := time.Now()
	return &` + pascal + `{
		CreatedAt: now,
		UpdatedAt: now,
	}
}
`
			return gen.writeFile(".go", content)
		},
	}
}

func NewMakeValueObjectCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:value-object", flag.ContinueOnError)
	return &SimpleCommand{
		use:     "make:value-object [name]",
		short:   "Create a new value object in internal/domain/valueobjects",
		argsMin: 1,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			gen := Generator{Name: name, Type: "Value Object", BasePath: "internal/domain/valueobjects", Package: "valueobjects"}
			content := `package valueobjects

import "fmt"

type ` + pascal + ` struct {
	value string
}

func New` + pascal + `(raw string) (` + pascal + `, error) {
	if raw == "" {
		return ` + pascal + `{}, fmt.Errorf("` + strings.ToLower(pascal) + ` cannot be empty")
	}
	return ` + pascal + `{value: raw}, nil
}

func Must` + pascal + `(raw string) ` + pascal + ` {
	v, err := New` + pascal + `(raw)
	if err != nil {
		panic(err)
	}
	return v
}

func (v ` + pascal + `) String() string { return v.value }
func (v ` + pascal + `) IsEmpty() bool  { return v.value == "" }
func (v ` + pascal + `) Equals(other ` + pascal + `) bool { return v.value == other.value }
`
			return gen.writeFile(".go", content)
		},
	}
}

func NewMakePortCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:port", flag.ContinueOnError)
	driving := fs.Bool("driving", false, "Generate as driving (primary/inbound) port")
	return &SimpleCommand{
		use:     "make:port [name]",
		short:   "Create a new port (interface) in internal/ports",
		argsMin: 1,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			portType := "driven"
			base := "internal/ports/driven"
			if *driving {
				portType = "driving"
				base = "internal/ports/driving"
			}
			gen := Generator{Name: name + "_port", Type: "Port (" + portType + ")", BasePath: base, Package: portType}
			mod := ModuleName()
			importBlock := `"context"`
			if *driving {
				importBlock = `"context"

	"` + mod + `/internal/domain/entities"
	"` + mod + `/internal/domain/valueobjects"`
			}
			content := `package ` + portType + `

import (
	` + importBlock + `
)

type ` + pascal + ` interface {
	Example(ctx context.Context) error
}
`
			return gen.writeFile(".go", content)
		},
	}
}

func NewMakeUsecaseCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:usecase", flag.ContinueOnError)
	return &SimpleCommand{
		use:     "make:usecase [name]",
		short:   "Create a new use case interactor in internal/application/usecases",
		argsMin: 1,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			suffix := "Interactor"
			if !strings.HasSuffix(strings.ToLower(pascal), "usecase") && !strings.HasSuffix(strings.ToLower(pascal), "interactor") {
				pascal += suffix
			}
			gen := Generator{Name: strings.TrimSuffix(pascal, suffix) + "_usecase", Type: "Use Case", BasePath: "internal/application/usecases", Package: "usecases"}
			mod := ModuleName()
			content := `package usecases

import (
	"context"

	"` + mod + `/internal/domain/entities"
	"` + mod + `/internal/domain/valueobjects"
	derrors "` + mod + `/internal/domain/errors"
	"` + mod + `/internal/ports/driven"
	"` + mod + `/internal/ports/driving"
)

type ` + pascal + ` struct {
}

// TODO: Inject driven ports (e.g. repo driven.XxxRepository) via constructor.
func New` + pascal + `() driving.XxxUseCase {
	return &` + pascal + `{}
}

func (uc *` + pascal + `) Example(ctx context.Context, id valueobjects.ID) (*entities.Entity, error) {
	_ = derrors.ErrNotFound
	return nil, nil
}
`
			return gen.writeFile(".go", content)
		},
	}
}

func NewMakeAdapterCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:adapter", flag.ContinueOnError)
	direction := fs.String("direction", "driven", "Adapter direction: driving (HTTP/gRPC/CLI) or driven (DB/Cache/Storage/External)")
	return &SimpleCommand{
		use:     "make:adapter [name]",
		short:   "Create a new infrastructure adapter in internal/adapters",
		argsMin: 1,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			snake := toSnake(name)
			dir := *direction
			if dir != "driving" {
				dir = "driven"
			}
			base := "internal/adapters/" + dir + "/" + snake
			gen := Generator{Name: snake + "_adapter", Type: "Adapter (" + dir + ")", BasePath: base, Package: snake}
			mod := ModuleName()
			content := `package ` + snake + `

import (
	"context"

	"` + mod + `/internal/ports/` + dir + `"
)

type ` + pascal + `Adapter struct {
}

func New` + pascal + `Adapter() ` + dir + `.` + pascal + ` {
	return &` + pascal + `Adapter{}
}

func (a *` + pascal + `Adapter) Example(ctx context.Context) error {
	return nil
}
`
			return gen.writeFile(".go", content)
		},
	}
}

func NewMakeModelCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:model", flag.ContinueOnError)
	withMigration := fs.Bool("migration", false, "Also generate a migration file")
	fs.BoolVar(withMigration, "m", false, "Also generate a migration file (short)")
	return &SimpleCommand{
		use:     "make:model [name]",
		short:   "Create a new persistence model (ORM-style)",
		argsMin: 1,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			gen := Generator{Name: name, Type: "Model", BasePath: "app/Models", Package: "Models"}
			table := toSnake(pascal) + "s"
			content := `package Models

import (
	"time"
	"gorm.io/gorm"
)

type ` + pascal + ` struct {
	ID        uint           ` + "`gorm:\"primaryKey\" json:\"id\"`" + `
	CreatedAt time.Time      ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time      ` + "`json:\"updated_at\"`" + `
	DeletedAt gorm.DeletedAt ` + "`gorm:\"index\" json:\"deleted_at,omitempty\"`" + `
}

func (` + pascal + `) TableName() string { return "` + table + `" }
`
			if err := gen.writeFile(".go", content); err != nil {
				return err
			}
			if *withMigration {
				return makeMigration("create_" + table + "_table")
			}
			return nil
		},
	}
}

func NewMakeMigrationCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:migration", flag.ContinueOnError)
	return &SimpleCommand{
		use:     "make:migration [name]",
		short:   "Create a new migration file in database/migrations",
		argsMin: 1,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			return makeMigration(args[0])
		},
	}
}

func makeMigration(name string) error {
	dir := "database/migrations"
	_ = os.MkdirAll(dir, 0755)
	ts := time.Now().Format("20060102150405")
	snake := toSnake(name)
	filename := fmt.Sprintf("%s_%s.go", ts, snake)
	path := dir + "/" + filename
	content := `package migrations

// Migration: ` + snake + `
// Generated at ` + time.Now().Format(time.RFC3339) + `
//
// Up applies the migration.
func Up() []string {
	return []string{
		"-- ` + snake + ` UP",
	}
}

// Down reverts the migration.
func Down() []string {
	return []string{
		"-- ` + snake + ` DOWN",
	}
}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✅ Created Migration: %s\n", path)
	return nil
}

func NewTinkerCommand() *SimpleCommand {
	fs := flag.NewFlagSet("tinker", flag.ContinueOnError)
	return &SimpleCommand{
		use:     "tinker",
		short:   "REPL to explore container, ports and architecture",
		argsMin: 0,
		argsMax: 0,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			fmt.Println(`
  🔮 Rancago Tinker REPL (minimal)
  Commands: help, ports, ls, info, quit
`)
			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Print("rancago> ")
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				line = strings.TrimSpace(line)
				switch line {
				case "":
					continue
				case "quit", "exit":
					fmt.Println("  Goodbye!")
					return nil
				case "help":
					fmt.Println("  Commands: help, ports, ls, info, quit")
				case "ports":
					fmt.Println("\n  Driving ports (inbound adapters implement these):")
					fmt.Println("    - NotificationUseCase")
					fmt.Println("    - UserUseCase")
					fmt.Println("    - DocumentUseCase")
					fmt.Println("\n  Driven ports (outbound adapters implement these):")
					fmt.Println("    - Repository[T] / VectorRepository[T]")
					fmt.Println("    - NotificationRepository / UserRepository / RoleRepository / PermissionRepository / DocumentRepository")
					fmt.Println("    - CachePort / StorageDriver / StorageManagerPort")
					fmt.Println("    - WebSocketPort / AuthProviderPort / SocialitePort")
					fmt.Println("    - DatabasePort / TransactionPort")
					fmt.Println()
				case "ls":
					fmt.Println("\n  Container bindings (kernel.Container):")
					fmt.Println("    - config           : *kernel.Config")
					fmt.Println("    - redis            : driven.CachePort")
					fmt.Println("    - ws.hub           : driven.WebSocketPort")
					fmt.Println("    - storage          : driven.StorageManagerPort")
					fmt.Println("    - socialite        : driven.SocialitePort")
					fmt.Println("    - repo.notification: driven.NotificationRepository")
					fmt.Println("    - repo.user        : driven.UserRepository")
					fmt.Println("    - repo.role        : driven.RoleRepository")
					fmt.Println("    - repo.permission  : driven.PermissionRepository")
					fmt.Println("    - repo.document    : driven.DocumentRepository")
					fmt.Println("    - uc.notification  : driving.NotificationUseCase")
					fmt.Println("    - uc.user          : driving.UserUseCase")
					fmt.Println("    - uc.document      : driving.DocumentUseCase")
					fmt.Println()
				case "info":
					fmt.Println("\n  Rancago Framework 1.0.0 - Hexagonal Architecture Edition")
					fmt.Println("  Go module: " + ModuleName())
					fmt.Println("  Pattern: Ports & Adapters (Hexagonal)")
					fmt.Println("  Layers:  domain → ports → application → adapters → bootstrap")
					fmt.Println()
				default:
					fmt.Printf("  unknown command: %q (type 'help')\n", line)
				}
			}
			return nil
		},
	}
}

func NewKeyGenerateCommand() *SimpleCommand {
	fs := flag.NewFlagSet("key:generate", flag.ContinueOnError)
	return &SimpleCommand{
		use:     "key:generate",
		short:   "Generate a new application key",
		argsMin: 0,
		argsMax: 0,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			b := make([]byte, 32)
			for i := range b {
				b[i] = byte((i*17 + 3) % 256)
			}
			key := fmt.Sprintf("base64:%x", b)
			fmt.Printf("  🔑 APP_KEY=%s\n", key)
			fmt.Println("  (Set this as environment variable APP_KEY)")
			return nil
		},
	}
}

func NewStorageLinkCommand() *SimpleCommand {
	fs := flag.NewFlagSet("storage:link", flag.ContinueOnError)
	return &SimpleCommand{
		use:     "storage:link",
		short:   "Create a symbolic link public/storage -> storage/app/public",
		argsMin: 0,
		argsMax: 0,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			_ = os.MkdirAll("storage/app/public", 0755)
			linkPath := "public/storage"
			if _, err := os.Lstat(linkPath); err == nil {
				fmt.Println("  ℹ️  Link already exists:", linkPath)
				return nil
			}
			if err := os.Symlink("../storage/app/public", linkPath); err != nil {
				return fmt.Errorf("could not create symlink (run as admin on Windows): %w", err)
			}
			fmt.Println("  ✅ Storage link created:", linkPath, "-> storage/app/public")
			return nil
		},
	}
}

func NewRouteListCommand() *SimpleCommand {
	fs := flag.NewFlagSet("route:list", flag.ContinueOnError)
	return &SimpleCommand{
		use:     "route:list",
		short:   "List registered HTTP and gRPC routes",
		argsMin: 0,
		argsMax: 0,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			fmt.Println("\n  Rancago Routes")
			fmt.Println("  " + stringsRepeat("=", 90))
			fmt.Printf("  %-8s %-50s %s\n", "METHOD", "PATH", "MIDDLEWARE / NOTES")
			fmt.Println("  " + stringsRepeat("-", 90))
			routes := []struct{ m, p, mw string }{
				{"GET", "/", "web"},
				{"GET", "/api/v1/health", "api"},
				{"POST", "/api/v1/notifications/send", "NotificationUseCase.Send"},
				{"POST", "/api/v1/notifications/broadcast", "NotificationUseCase.Broadcast"},
				{"GET", "/api/v1/notifications/list", "NotificationUseCase.ListUserNotifications"},
				{"GET", "/api/v1/notifications/count", "NotificationUseCase.GetUnreadCount"},
				{"POST", "/api/v1/notifications/read", "NotificationUseCase.MarkRead"},
				{"GET", "/ws", "HubAdapter (WS stub)"},
				{"gRPC", "/gawego.NotificationService/*", "GRPCNotificationAdapter (stub)"},
			}
			for _, r := range routes {
				fmt.Printf("  %-8s %-50s %s\n", r.m, r.p, r.mw)
			}
			fmt.Println()
			return nil
		},
	}
}

type scaffoldAsk struct {
	Name      string
	HasEntity bool
	HasRepo   bool
	HasUC     bool
	HasHTTP   bool
	HasGRPC   bool
}

func runSub(cmd *SimpleCommand, argStr string) error {
	cmd.SetArgs(strings.Fields(argStr))
	return cmd.Execute()
}

func NewScaffoldCommand() *SimpleCommand {
	fs := flag.NewFlagSet("scaffold", flag.ContinueOnError)
	name := fs.String("name", "", "Component name (also accepted as positional arg)")
	noAsk := fs.Bool("no-interactive", false, "Disable prompts; use explicit flags")
	entity := fs.Bool("entity", true, "Scaffold entity (with --no-interactive)")
	repo := fs.Bool("repo", true, "Scaffold repository driven port (with --no-interactive)")
	usecase := fs.Bool("usecase", true, "Scaffold usecase + port (with --no-interactive)")
	http := fs.Bool("http", true, "Scaffold HTTP handler adapter (with --no-interactive)")
	grpc := fs.Bool("grpc", false, "Scaffold gRPC adapter stub (with --no-interactive)")
	fs.StringVar(name, "n", "", "Component name (short)")
	return &SimpleCommand{
		use:     "scaffold [name]",
		short:   "Interactive scaffolder: entity + port + usecase + adapter",
		argsMin: 0,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			_ = context.Background()
			ask := scaffoldAsk{Name: *name}
			if len(args) > 0 && ask.Name == "" {
				ask.Name = args[0]
			}
			if !*noAsk {
				r := bufio.NewReader(os.Stdin)
				if ask.Name == "" {
					ask.Name = promptLine(r, "  🏗️  Component name (e.g. Order, Project, Invoice): ")
				}
				ask.HasEntity = promptBool(r, "  → Create domain entity? [Y/n] ", true)
				ask.HasRepo = promptBool(r, "  → Create driven repository port? [Y/n] ", true)
				ask.HasUC = promptBool(r, "  → Create driving use case port + interactor? [Y/n] ", true)
				ask.HasHTTP = promptBool(r, "  → Create HTTP driving adapter (handler)? [Y/n] ", true)
				ask.HasGRPC = promptBool(r, "  → Create gRPC driving adapter (stub)? [y/N] ", false)
			} else {
				ask.HasEntity = *entity
				ask.HasRepo = *repo
				ask.HasUC = *usecase
				ask.HasHTTP = *http
				ask.HasGRPC = *grpc
			}
			if ask.Name == "" {
				return fmt.Errorf("component name is required (pass as positional arg or --name)")
			}
			fmt.Printf("\n  🚀 Scaffolding bounded context %q (hexagonal)\n", ask.Name)
			fmt.Println("  " + stringsRepeat("=", 60))
			if ask.HasEntity {
				if err := runSub(NewMakeEntityCommand(), ask.Name); err != nil {
					return err
				}
			}
			if ask.HasRepo {
				if err := runSub(NewMakePortCommand(), ask.Name+"Repository"); err != nil {
					return err
				}
			}
			if ask.HasUC {
				if err := runSub(NewMakePortCommand(), ask.Name+"UseCase --driving"); err != nil {
					return err
				}
				if err := runSub(NewMakeUsecaseCommand(), ask.Name); err != nil {
					return err
				}
			}
			if ask.HasHTTP {
				if err := runSub(NewMakeAdapterCommand(), ask.Name+"Handler --direction driving"); err != nil {
					return err
				}
			}
			if ask.HasGRPC {
				if err := runSub(NewMakeAdapterCommand(), ask.Name+"Grpc --direction driving"); err != nil {
					return err
				}
			}
			fmt.Printf("\n  ✅ Scaffold %q done! Wire adapters in internal/bootstrap and you're live.\n", ask.Name)

			// Auto-generate FEATURE.md for AI vibe coding
			var generated []generatedFile
			if ask.HasEntity {
				generated = append(generated, generatedFile{
					Layer: "Domain Entity",
					Path:  "internal/domain/entities/" + toPascal(ask.Name) + ".go",
					Role:  "Core business object with state and behavior",
				})
			}
			if ask.HasRepo {
				generated = append(generated, generatedFile{
					Layer: "Driven Port (Repository interface)",
					Path:  "internal/ports/driven/" + toPascal(ask.Name) + "Repository.go",
					Role:  "Outbound persistence contract",
				})
			}
			if ask.HasUC {
				generated = append(generated, generatedFile{
					Layer: "Driving Port (Use Case interface)",
					Path:  "internal/ports/driving/" + toPascal(ask.Name) + "UseCase.go",
					Role:  "Inbound contract - what HTTP/gRPC/CLI can call",
				})
				generated = append(generated, generatedFile{
					Layer: "Application Use Case (Interactor)",
					Path:  "internal/application/usecases/" + toSnake(ask.Name) + "_usecase.go",
					Role:  "Business logic - orchestrates domain + driven ports",
				})
			}
			if ask.HasHTTP {
				generated = append(generated, generatedFile{
					Layer: "Driving Adapter (HTTP Handler)",
					Path:  "internal/adapters/driving/" + toSnake(ask.Name) + "handler/" + toSnake(ask.Name) + "handler_adapter.go",
					Role:  "HTTP entry point",
				})
			}
			if ask.HasGRPC {
				generated = append(generated, generatedFile{
					Layer: "Driving Adapter (gRPC)",
					Path:  "internal/adapters/driving/" + toSnake(ask.Name) + "grpc/" + toSnake(ask.Name) + "grpc_adapter.go",
					Role:  "gRPC entry point stub",
				})
			}
			mdPath, err := writeFeatureMarkdown(ask.Name, "", generated, ask)
			if err != nil {
				fmt.Printf("  ⚠️  Could not write FEATURE.md: %v\n", err)
			} else {
				fmt.Printf("  📄 AI context: %s\n", mdPath)
			}
			return nil
		},
	}
}
