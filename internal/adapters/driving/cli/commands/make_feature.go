package commands

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// NewMakeFeatureCommand generates a complete feature (scaffold) + FEATURE.md
// for vibe-coding with any AI.
func NewMakeFeatureCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:feature", flag.ContinueOnError)
	noInteractive := fs.Bool("no-interactive", false, "Skip prompts; use explicit flags")
	entity := fs.Bool("entity", true, "Generate domain entity")
	repo := fs.Bool("repo", true, "Generate driven repository port")
	usecase := fs.Bool("usecase", true, "Generate driving use case port + interactor")
	httpAdapter := fs.Bool("http", true, "Generate HTTP driving adapter")
	grpcAdapter := fs.Bool("grpc", false, "Generate gRPC driving adapter")
	desc := fs.String("desc", "", "Short description of the feature (for FEATURE.md)")

	return &SimpleCommand{
		use:     "make:feature [name]",
		short:   "Scaffold a feature + generate FEATURE.md for AI-assisted vibe coding",
		argsMin: 0,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			ask := scaffoldAsk{}
			if len(args) > 0 {
				ask.Name = args[0]
			}

			if !*noInteractive {
				r := bufio.NewReader(os.Stdin)
				if ask.Name == "" {
					ask.Name = promptLine(r, "  🏗️  Feature name (e.g. Order, Invoice, Product): ")
				}
				if *desc == "" {
					*desc = promptLine(r, "  📝 One-line description (for FEATURE.md): ")
				}
				ask.HasEntity = promptBool(r, "  → Create domain entity? [Y/n] ", true)
				ask.HasRepo = promptBool(r, "  → Create driven repository port? [Y/n] ", true)
				ask.HasUC = promptBool(r, "  → Create driving use case port + interactor? [Y/n] ", true)
				ask.HasHTTP = promptBool(r, "  → Create HTTP driving adapter (handler)? [Y/n] ", true)
				ask.HasGRPC = promptBool(r, "  → Create gRPC driving adapter stub? [y/N] ", false)
			} else {
				ask.HasEntity = *entity
				ask.HasRepo = *repo
				ask.HasUC = *usecase
				ask.HasHTTP = *httpAdapter
				ask.HasGRPC = *grpcAdapter
			}

			if ask.Name == "" {
				return fmt.Errorf("feature name is required")
			}

			fmt.Printf("\n  🚀 Scaffolding feature %q (hexagonal)\n", ask.Name)
			fmt.Println("  " + stringsRepeat("=", 60))

			// Track which files were generated
			var generated []generatedFile

			if ask.HasEntity {
				if err := runSub(NewMakeEntityCommand(), ask.Name); err != nil {
					return err
				}
				generated = append(generated, generatedFile{
					Layer: "Domain Entity",
					Path:  "internal/domain/entities/" + toPascal(ask.Name) + ".go",
					Role:  "Core business object with state and behavior",
				})
			}
			if ask.HasRepo {
				if err := runSub(NewMakePortCommand(), ask.Name+"Repository"); err != nil {
					return err
				}
				generated = append(generated, generatedFile{
					Layer: "Driven Port (Repository interface)",
					Path:  "internal/ports/driven/" + toPascal(ask.Name) + "Repository.go",
					Role:  "Outbound persistence contract - swap DB without touching business logic",
				})
			}
			if ask.HasUC {
				if err := runSub(NewMakePortCommand(), ask.Name+"UseCase --driving"); err != nil {
					return err
				}
				generated = append(generated, generatedFile{
					Layer: "Driving Port (Use Case interface)",
					Path:  "internal/ports/driving/" + toPascal(ask.Name) + "UseCase.go",
					Role:  "Inbound contract - what HTTP/gRPC/CLI can call",
				})
				if err := runSub(NewMakeUsecaseCommand(), ask.Name); err != nil {
					return err
				}
				generated = append(generated, generatedFile{
					Layer: "Application Use Case (Interactor)",
					Path:  "internal/application/usecases/" + toSnake(ask.Name) + "_usecase.go",
					Role:  "Business logic - orchestrates domain + driven ports",
				})
			}
			if ask.HasHTTP {
				if err := runSub(NewMakeAdapterCommand(), ask.Name+"Handler --direction driving"); err != nil {
					return err
				}
				generated = append(generated, generatedFile{
					Layer: "Driving Adapter (HTTP Handler)",
					Path:  "internal/adapters/driving/" + toSnake(ask.Name) + "handler/" + toSnake(ask.Name) + "handler_adapter.go",
					Role:  "HTTP entry point - decode request, call use case, encode response",
				})
			}
			if ask.HasGRPC {
				if err := runSub(NewMakeAdapterCommand(), ask.Name+"Grpc --direction driving"); err != nil {
					return err
				}
				generated = append(generated, generatedFile{
					Layer: "Driving Adapter (gRPC)",
					Path:  "internal/adapters/driving/" + toSnake(ask.Name) + "grpc/" + toSnake(ask.Name) + "grpc_adapter.go",
					Role:  "gRPC entry point stub",
				})
			}

			// Generate FEATURE.md
			mdPath, err := writeFeatureMarkdown(ask.Name, *desc, generated, ask)
			if err != nil {
				return fmt.Errorf("could not write FEATURE.md: %w", err)
			}

			fmt.Printf("\n  ✅ Feature %q scaffolded!\n", ask.Name)
			fmt.Printf("  📄 AI context file: %s\n", mdPath)
			fmt.Printf("  📌 Wire adapters in internal/bootstrap/app.go to go live.\n\n")
			return nil
		},
	}
}

type generatedFile struct {
	Layer string
	Path  string
	Role  string
}

func writeFeatureMarkdown(name, desc string, files []generatedFile, ask scaffoldAsk) (string, error) {
	pascal := toPascal(name)
	snake := toSnake(name)
	dir := "docs/features"
	_ = os.MkdirAll(dir, 0755)
	path := dir + "/" + snake + ".md"
	mod := ModuleName()

	if desc == "" {
		desc = pascal + " bounded context"
	}

	var sb strings.Builder

	// ── CONTEXT block ────────────────────────────────────────────────────────
	sb.WriteString("<!-- CONTEXT_START\n")
	sb.WriteString("module: " + mod + "\n")
	sb.WriteString("feature: " + pascal + "\n")
	sb.WriteString("generated: " + time.Now().Format("2006-01-02") + "\n")
	sb.WriteString("arch: hexagonal (ports-and-adapters)\n")
	sb.WriteString("CONTEXT_END -->\n\n")

	sb.WriteString("# Feature: " + pascal + "\n\n")
	sb.WriteString("> " + desc + "\n\n")

	// ── INSTRUCTION block (what to do when working on this feature) ─────────
	sb.WriteString("## 📋 Instructions\n\n")
	sb.WriteString("<!-- INSTRUCTION\n")
	sb.WriteString("Read this file FIRST before any edit. Follow hexagonal rules:\n")
	sb.WriteString("1. Domain entities MUST NOT import ports or adapters.\n")
	sb.WriteString("2. Ports are Go interfaces only - no implementations.\n")
	sb.WriteString("3. Use cases depend on driven ports (injected via constructor).\n")
	sb.WriteString("4. Adapters depend on driving ports - never on use case structs directly.\n")
	sb.WriteString("5. Wire new bindings in internal/bootstrap/app.go (Container.Singleton).\n")
	sb.WriteString("6. Use derrors.New(op, sentinel, msg) for domain errors.\n")
	sb.WriteString("7. IDs use valueobjects.ID - call valueobjects.NewIDStr() or NewIDUint().\n")
	sb.WriteString("INSTRUCTION -->\n\n")

	// ── Files ────────────────────────────────────────────────────────────────
	sb.WriteString("## 📁 Generated Files\n\n")
	sb.WriteString("| Layer | File | Role |\n")
	sb.WriteString("|-------|------|------|\n")
	for _, f := range files {
		sb.WriteString("| " + f.Layer + " | `" + f.Path + "` | " + f.Role + " |\n")
	}
	sb.WriteString("\n")

	// ── Architecture diagram (compact) ──────────────────────────────────────
	sb.WriteString("## 🏗️ Layer Flow\n\n")
	sb.WriteString("```\n")
	sb.WriteString("HTTP/gRPC/CLI\n")
	sb.WriteString("  └─ Driving Adapter  (internal/adapters/driving/)\n")
	sb.WriteString("       └─ Driving Port   (internal/ports/driving/" + pascal + "UseCase)\n")
	sb.WriteString("            └─ Use Case      (internal/application/usecases/" + snake + "_usecase.go)\n")
	sb.WriteString("                 └─ Driven Port   (internal/ports/driven/" + pascal + "Repository)\n")
	sb.WriteString("                      └─ Adapter      (internal/adapters/driven/persistence/)\n")
	sb.WriteString("                           └─ Domain       (internal/domain/entities/" + pascal + ".go)\n")
	sb.WriteString("```\n\n")

	// ── Bootstrap wiring ─────────────────────────────────────────────────────
	sb.WriteString("## 🔌 Bootstrap Wiring\n\n")
	sb.WriteString("Add to `internal/bootstrap/app.go` inside `RegisterCore()`:\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("// " + pascal + " - repository\n")
	sb.WriteString("a.Container.Singleton(\"repo." + toSnake(name) + "\", func(c *kernel.Container) (interface{}, error) {\n")
	sb.WriteString("    return inmemory.NewInMemory" + pascal + "Repo(), nil\n")
	sb.WriteString("})\n")
	if ask.HasUC {
		sb.WriteString("\n// " + pascal + " - use case\n")
		sb.WriteString("a.Container.Singleton(\"uc." + toSnake(name) + "\", func(c *kernel.Container) (interface{}, error) {\n")
		sb.WriteString("    repoRaw, _ := c.Resolve(\"repo." + toSnake(name) + "\")\n")
		sb.WriteString("    return usecases.New" + pascal + "Interactor(\n")
		sb.WriteString("        repoRaw.(driven." + pascal + "Repository),\n")
		sb.WriteString("    ), nil\n")
		sb.WriteString("})\n")
		sb.WriteString("a.Container.Alias(\"uc." + toSnake(name) + "\", \"driving." + pascal + "UseCase\")\n")
	}
	if ask.HasHTTP {
		sb.WriteString("\n// " + pascal + " - HTTP handler\n")
		sb.WriteString("// In BuildHTTPServer():\n")
		sb.WriteString("ucRaw, _ := a.Container.Resolve(\"uc." + toSnake(name) + "\")\n")
		sb.WriteString("nh := " + toSnake(name) + "handler.New" + pascal + "HandlerAdapter(ucRaw.(driving." + pascal + "UseCase))\n")
		sb.WriteString("nh.RegisterRoutes(mux, \"/api/v1/" + toSnake(name) + "s\")\n")
	}
	sb.WriteString("```\n\n")

	// ── Common tasks (OUTPUT hints) ───────────────────────────────────────────
	sb.WriteString("## ⚡ Quick Tasks\n\n")
	sb.WriteString("<!-- OUTPUT_HINTS\n")
	sb.WriteString("When asked to add a method to this feature:\n")
	sb.WriteString("  1. Add method signature to internal/ports/driving/" + pascal + "UseCase.go\n")
	sb.WriteString("  2. Implement in internal/application/usecases/" + snake + "_usecase.go\n")
	sb.WriteString("  3. Add HTTP handler in the driving adapter RegisterRoutes()\n")
	sb.WriteString("  4. Update this file's Quick Tasks table\n\n")
	sb.WriteString("When asked to add a field to the entity:\n")
	sb.WriteString("  1. Edit internal/domain/entities/" + pascal + ".go\n")
	sb.WriteString("  2. Update New" + pascal + "() constructor\n")
	sb.WriteString("  3. Create a migration: rancago make:migration add_field_to_" + snake + "s\n")
	sb.WriteString("OUTPUT_HINTS -->\n\n")

	sb.WriteString("| Task | Command |\n")
	sb.WriteString("|------|---------|\n")
	sb.WriteString("| Add field to entity | Edit `internal/domain/entities/" + pascal + ".go` |\n")
	sb.WriteString("| Add use case method | Edit port + interactor |\n")
	sb.WriteString("| Add HTTP route | Edit driving adapter `RegisterRoutes()` |\n")
	sb.WriteString("| Add migration | `rancago make:migration add_..._to_" + snake + "s` |\n")
	sb.WriteString("| Add repository method | Edit `internal/ports/driven/" + pascal + "Repository.go` |\n")
	sb.WriteString("\n")

	// ── Errors reference ─────────────────────────────────────────────────────
	sb.WriteString("## 🚨 Domain Errors\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("// Available sentinels (internal/domain/errors/errors.go)\n")
	sb.WriteString("derrors.ErrNotFound      // 404-equivalent\n")
	sb.WriteString("derrors.ErrUnauthorized  // 401\n")
	sb.WriteString("derrors.ErrForbidden     // 403\n")
	sb.WriteString("derrors.ErrValidation    // 422\n")
	sb.WriteString("derrors.ErrConflict      // 409\n")
	sb.WriteString("derrors.ErrAlreadyExists // 409 (unique)\n\n")
	sb.WriteString("// Wrap with operation context:\n")
	sb.WriteString("return derrors.New(\"" + snake + ".create\", derrors.ErrValidation, \"name is required\")\n")
	sb.WriteString("```\n")

	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return "", err
	}
	fmt.Printf("  📄 Created Feature Docs: %s\n", path)
	return path, nil
}
