package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rancago/framework/internal/adapters/driving/cli/commands"
	"github.com/rancago/framework/internal/bootstrap"
)

const (
	Version   = "1.0.0"
	BuildDate = "2026-07-27"
)

type command struct {
	name  string
	short string
	long  string
	run   func(args []string) error
}

func rancagoBanner() string {
	return `
  ____                                        
 |  _ \ __ _ _ __   ___ __ _  __ _  ___  
 | |_) / _` + "`" + ` | '_ \ / __/ _` + "`" + ` |/ _` + "`" + ` |/ _ \ 
 |  _ < (_| | | | | (_| (_| | (_| | (_) |
 |_| \_\__,_|_| |_|\___\__,_|\__, |\___/ 
                               |___/       
`
}

func RunCLI() int {
	if len(os.Args) < 2 {
		printUsage()
		return 0
	}
	cmdName := os.Args[1]
	args := os.Args[2:]
	switch cmdName {
	case "-h", "--help", "help":
		printUsage()
		return 0
	case "-v", "--version", "version":
		fmt.Printf("Rancago Framework %s (built %s)\n", Version, BuildDate)
		return 0
	case "serve":
		return handleServe(args)
	case "migrate":
		return runCmd(commands.NewMigrateCommand(), args)
	case "scaffold":
		return runCmd(commands.NewScaffoldCommand(), args)
	case "make:entity":
		return runCmd(commands.NewMakeEntityCommand(), args)
	case "make:value-object", "make:vo":
		return runCmd(commands.NewMakeValueObjectCommand(), args)
	case "make:port":
		return runCmd(commands.NewMakePortCommand(), args)
	case "make:usecase":
		return runCmd(commands.NewMakeUsecaseCommand(), args)
	case "make:adapter":
		return runCmd(commands.NewMakeAdapterCommand(), args)
	case "make:model":
		return runCmd(commands.NewMakeModelCommand(), args)
	case "make:migration":
		return runCmd(commands.NewMakeMigrationCommand(), args)
	case "tinker":
		return runCmd(commands.NewTinkerCommand(), args)
	case "key:generate", "key:gen":
		return runCmd(commands.NewKeyGenerateCommand(), args)
	case "storage:link":
		return runCmd(commands.NewStorageLinkCommand(), args)
	case "route:list":
		return runCmd(commands.NewRouteListCommand(), args)
	default:
		if strings.HasPrefix(cmdName, "make:") {
			fmt.Fprintf(os.Stderr, "Unknown make command: %s\n", cmdName)
			fmt.Fprintln(os.Stderr, "Available: make:entity, make:value-object, make:port, make:usecase, make:adapter, make:model, make:migration")
			return 1
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %s (try: rancago help)\n", cmdName)
		return 1
	}
}

func printUsage() {
	fmt.Println(rancagoBanner())
	fmt.Println("Rancago Framework CLI - Clean Hexagonal Architecture Toolkit")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  rancago [command] [flags] [args]")
	fmt.Println()
	fmt.Println("Commands:")
	rows := [][2]string{
		{"serve", "Start the HTTP (and optional gRPC) development server"},
		{"migrate", "Run database migrations (stub)"},
		{"scaffold [name]", "Interactive scaffolder for a bounded context"},
		{"", ""},
		{"Code generation:", ""},
		{"  make:entity [name]", "Create a domain entity"},
		{"  make:value-object [name]", "Create a value object"},
		{"  make:port [name]", "Create a driving/driven port interface"},
		{"  make:usecase [name]", "Create a use case interactor"},
		{"  make:adapter [name]", "Create a driving/driven adapter"},
		{"  make:model [name] [-m]", "Create an ORM model (with optional migration)"},
		{"  make:migration [name]", "Create a migration file"},
		{"", ""},
		{"Utility:", ""},
		{"  tinker", "Explore container / architecture via REPL"},
		{"  key:generate", "Generate a new APP_KEY"},
		{"  storage:link", "Symlink public/storage -> storage/app/public"},
		{"  route:list", "Show registered HTTP and gRPC routes"},
		{"", ""},
		{"  help", "Show this help"},
		{"  version / -v", "Show framework version"},
	}
	for _, r := range rows {
		if r[0] == "" {
			fmt.Println(r[1])
		} else {
			fmt.Printf("  %-28s %s\n", r[0], r[1])
		}
	}
	fmt.Println()
	fmt.Println("Run 'rancago <command> --help' for command-specific flags.")
}

func handleServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 0, "HTTP server port (overrides config)")
	host := fs.String("host", "0.0.0.0", "Server host (binding hint)")
	withGRPC := fs.Bool("grpc", false, "Also start gRPC stub server alongside HTTP")
	fs.IntVar(port, "p", 0, "HTTP server port (short)")
	fs.StringVar(host, "H", "0.0.0.0", "Server host (short)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	app := bootstrap.New()
	app.RegisterCore()
	app.Boot()
	if *port != 0 {
		app.Config.Server.HTTPPort = *port
	}
	if *host != "" && *host != "0.0.0.0" {
		fmt.Printf("[rancago] Host hint: %s (stub server binds to all interfaces)\n", *host)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	errCh := make(chan error, 2)
	go func() { errCh <- app.StartHTTP(ctx) }()
	if *withGRPC {
		go func() { errCh <- app.StartGRPC(ctx) }()
	}
	if err := <-errCh; err != nil {
		fmt.Fprintf(os.Stderr, "[rancago] server error: %v\n", err)
		return 1
	}
	return 0
}

type cobralike interface {
	SetArgs([]string)
	Execute() error
}

func runCmd(cmd cobralike, args []string) int {
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
