# CLAUDE.md - AI Assistant Guide for Facienda

**Last Updated**: 2025-11-15
**Project**: Facienda - Console TODO App in Go
**License**: MIT
**Author**: John Mirolha

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Current State](#current-state)
3. [Technology Stack](#technology-stack)
4. [Expected Project Structure](#expected-project-structure)
5. [Development Workflows](#development-workflows)
6. [Go Conventions & Best Practices](#go-conventions--best-practices)
7. [Testing Standards](#testing-standards)
8. [Git Workflow](#git-workflow)
9. [AI Assistant Guidelines](#ai-assistant-guidelines)
10. [Common Tasks](#common-tasks)

---

## Project Overview

**Facienda** is a console-based TODO application written in Go. The project aims to provide a simple, efficient command-line interface for managing tasks and to-do items.

### Project Goals
- Provide a clean CLI interface for task management
- Demonstrate Go best practices and idiomatic code
- Maintain simplicity while offering essential TODO functionality
- Support common TODO operations (add, list, complete, delete, edit)

---

## Current State

**Status**: Early Development / Initial Setup

The repository is currently in its initial state with:
- Basic project files (README.md, LICENSE, .gitignore)
- Git repository initialized
- No source code implemented yet

**Next Steps**:
- Set up Go module initialization
- Create basic project structure
- Implement core TODO functionality
- Add CLI interface
- Implement data persistence

---

## Technology Stack

### Primary Technologies
- **Language**: Go (Golang)
- **Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
- **Build Tool**: Go toolchain (go build, go test)

### Expected Dependencies
When implementing features, prefer standard library solutions. Consider these libraries for specific needs:

- **CLI Framework**:
  - `github.com/spf13/cobra` - for robust CLI commands
  - `github.com/urfave/cli` - alternative CLI framework
  - Standard library `flag` package for simple cases

- **Data Persistence**:
  - JSON files (using standard library)
  - SQLite (`github.com/mattn/go-sqlite3`) for structured storage
  - Plain text formats for simplicity

- **Testing**:
  - Standard library `testing` package
  - `github.com/stretchr/testify` for assertions (optional)

---

## Expected Project Structure

As the project develops, follow this recommended structure:

```
facienda/
├── .git/                   # Git repository
├── .gitignore              # Go-specific gitignore
├── LICENSE                 # MIT License
├── README.md               # User-facing documentation
├── CLAUDE.md               # This file - AI assistant guide
├── go.mod                  # Go module file
├── go.sum                  # Go dependencies checksum
├── main.go                 # Application entry point
├── cmd/                    # Command implementations
│   ├── root.go            # Root command setup
│   ├── add.go             # Add task command
│   ├── list.go            # List tasks command
│   ├── complete.go        # Complete task command
│   └── delete.go          # Delete task command
├── internal/              # Private application code
│   ├── todo/              # TODO business logic
│   │   ├── todo.go        # Todo struct and methods
│   │   └── todo_test.go   # Unit tests
│   ├── storage/           # Data persistence layer
│   │   ├── storage.go     # Storage interface
│   │   ├── json.go        # JSON storage implementation
│   │   └── storage_test.go
│   └── config/            # Configuration handling
│       └── config.go
├── pkg/                   # Public, reusable packages (if needed)
├── testdata/              # Test fixtures and data
└── docs/                  # Additional documentation
```

### Directory Purposes

- **cmd/**: Command-line interface commands and subcommands
- **internal/**: Private application code that cannot be imported by other projects
- **pkg/**: Public libraries that can be imported (use sparingly)
- **testdata/**: Test fixtures, sample data files

---

## Development Workflows

### Initial Setup

```bash
# Initialize Go module (if not done)
go mod init github.com/johnmirolha/facienda

# Download dependencies
go mod download

# Tidy up dependencies
go mod tidy
```

### Building

```bash
# Build the application
go build -o facienda .

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o facienda-linux .
GOOS=windows GOARCH=amd64 go build -o facienda.exe .
GOOS=darwin GOARCH=amd64 go build -o facienda-mac .
```

### Running

```bash
# Run without building
go run main.go

# Run with arguments
go run main.go add "My new task"

# Run built binary
./facienda list
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests verbosely
go test -v ./...

# Run specific test
go test -run TestTodoAdd ./internal/todo
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code for common issues
go vet ./...

# Run static analysis (install golangci-lint first)
golangci-lint run

# Check for security issues (install gosec first)
gosec ./...
```

---

## Go Conventions & Best Practices

### Code Style

1. **Formatting**: Always use `go fmt` - code must be formatted before committing
2. **Naming Conventions**:
   - Use `MixedCaps` or `mixedCaps` (camelCase) - never underscores
   - Exported names start with capital letter
   - Package names are lowercase, single word
   - Interface names: single method interfaces end in "-er" (e.g., `Reader`, `Writer`)

3. **Error Handling**:
   ```go
   // Good: Check errors immediately
   if err != nil {
       return fmt.Errorf("failed to add task: %w", err)
   }

   // Bad: Ignoring errors
   data, _ := ioutil.ReadFile("file.txt")
   ```

4. **Comments**:
   - Exported functions/types must have doc comments
   - Doc comments start with the name of the element
   ```go
   // Add creates a new task with the given description.
   func Add(description string) error {
   ```

### Package Organization

- Keep packages focused and cohesive
- Avoid circular dependencies
- Use `internal/` for private packages
- Minimize dependencies between packages

### Idiomatic Go

1. **Accept interfaces, return structs**
   ```go
   func ProcessTodos(storage Storage) error {
       // Storage is an interface
   }
   ```

2. **Use contexts for cancellation and timeouts**
   ```go
   func FetchTodos(ctx context.Context) ([]Todo, error) {
   ```

3. **Prefer composition over inheritance**
   ```go
   type TodoList struct {
       items []Todo
       storage Storage
   }
   ```

4. **Keep the happy path left-aligned**
   ```go
   // Good
   if err != nil {
       return err
   }
   // success path continues

   // Avoid deep nesting
   ```

### Error Handling Patterns

```go
// Wrap errors with context
return fmt.Errorf("failed to save todo: %w", err)

// Define custom error types when needed
var ErrNotFound = errors.New("todo not found")

// Sentinel errors for comparison
if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

---

## Testing Standards

### Unit Tests

- Test files end with `_test.go`
- Test functions start with `Test`
- Use table-driven tests for multiple cases

```go
func TestTodoAdd(t *testing.T) {
    tests := []struct {
        name        string
        description string
        wantErr     bool
    }{
        {
            name:        "valid task",
            description: "Buy groceries",
            wantErr:     false,
        },
        {
            name:        "empty description",
            description: "",
            wantErr:     true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := AddTodo(tt.description)
            if (err != nil) != tt.wantErr {
                t.Errorf("AddTodo() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Coverage Goals

- Aim for >80% code coverage for critical paths
- Focus on business logic in `internal/`
- Don't obsess over 100% coverage

### Test Organization

- Unit tests: Test individual functions/methods
- Integration tests: Test component interactions
- Use `testdata/` for test fixtures

---

## Git Workflow

### Branch Naming

- Feature branches: `feature/description` or `claude/session-id`
- Bug fixes: `fix/description`
- Main development branch: `main` or `master`

### Commit Messages

Follow conventional commit format:

```
type(scope): subject

body (optional)

footer (optional)
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(todo): add task completion functionality

fix(storage): handle missing file gracefully

docs(readme): update installation instructions
```

### Before Committing

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Vet code
go vet ./...

# Stage and commit
git add .
git commit -m "feat(todo): implement add command"
```

---

## AI Assistant Guidelines

### When Adding Features

1. **Check existing code first**: Read relevant files before implementing
2. **Follow Go conventions**: Use idiomatic Go patterns
3. **Write tests**: Add unit tests for new functionality
4. **Update documentation**: Keep README.md and comments current
5. **Handle errors properly**: Never ignore errors
6. **Use standard library first**: Avoid unnecessary dependencies

### Code Quality Checklist

Before completing a task, verify:

- [ ] Code is formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] All tests pass (`go test ./...`)
- [ ] Error handling is comprehensive
- [ ] Public functions have doc comments
- [ ] Code follows Go naming conventions
- [ ] No sensitive data in code or commits

### Common Patterns

**Adding a new TODO command**:
1. Create command file in `cmd/` (e.g., `cmd/complete.go`)
2. Implement business logic in `internal/todo/`
3. Add tests in `internal/todo/todo_test.go`
4. Update storage if needed
5. Register command in `cmd/root.go`

**Adding data persistence**:
1. Define interface in `internal/storage/storage.go`
2. Implement concrete type (e.g., `json.go`)
3. Add tests for storage operations
4. Handle file not found and corruption gracefully

### Security Considerations

- Validate all user input
- Sanitize file paths to prevent directory traversal
- Handle file permissions appropriately
- Don't expose sensitive information in errors
- Validate JSON/data before unmarshaling

### Performance Guidelines

- Avoid premature optimization
- Use benchmarks for performance-critical code
- Profile before optimizing (use `pprof`)
- Keep memory allocations reasonable
- Use appropriate data structures

---

## Common Tasks

### Initializing the Project

```bash
# Create Go module
go mod init github.com/johnmirolha/facienda

# Create basic main.go
cat > main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Facienda TODO App")
}
EOF

# Test it works
go run main.go
```

### Adding a Dependency

```bash
# Add dependency (example: cobra CLI)
go get github.com/spf13/cobra@latest

# Import in code, then run
go mod tidy
```

### Creating a Release Build

```bash
# Build optimized binary
go build -ldflags="-s -w" -o facienda .

# The -ldflags="-s -w" strips debug info for smaller binary
```

### Debugging

```bash
# Run with race detector
go run -race main.go

# Use delve debugger
dlv debug main.go
```

---

## Resources

### Go Learning Resources
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)

### Tools
- `go fmt` - Code formatting
- `go vet` - Static analysis
- `golangci-lint` - Comprehensive linter
- `gopls` - Go language server
- `delve` - Go debugger

### Similar Projects for Reference
- [todo-cli](https://github.com/simonwhitaker/todo-cli)
- [task](https://github.com/go-task/task)
- [todoist-cli](https://github.com/sachaos/todoist)

---

## Questions?

For questions about this guide or the project:
- Check the [README.md](./README.md) for user documentation
- Review Go documentation at [golang.org/doc](https://golang.org/doc/)
- Examine existing code for patterns and conventions

---

**Remember**: Write clear, idiomatic Go code. When in doubt, prefer simplicity and readability over cleverness.
