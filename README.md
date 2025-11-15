# Facienda

A simple, efficient console-based TODO application written in Go.

## Features

- Add tasks with specific dates (defaults to today)
- Mark tasks as completed or incomplete
- Tasks include title and optional details
- Edit task details
- View current, past, and future tasks
- SQLite storage for persistence
- Cross-platform support (Linux, macOS, Windows)

## Installation

### From Source

```bash
git clone https://github.com/johnmirolha/facienda.git
cd facienda
go build -o facienda .
```

### Quick Install

```bash
go install github.com/johnmirolha/facienda@latest
```

## Usage

### Add a Task

```bash
# Add task for today
facienda add "Buy groceries"

# Add task with details
facienda add "Buy groceries" -m "Milk, eggs, bread"

# Add task for specific date
facienda add "Doctor appointment" -d 2025-12-15

# Combine date and details
facienda add "Meeting" -d 2025-11-20 -m "Discuss Q4 results"
```

### View Tasks

```bash
# View today's tasks
facienda list

# View past tasks (timeline)
facienda past

# View future tasks
facienda future
```

### Manage Tasks

```bash
# Mark task as completed
facienda complete 1

# Mark task as incomplete
facienda incomplete 1

# Edit task title
facienda edit 1 -t "New title"

# Edit task details
facienda edit 1 -m "New details"

# Edit both title and details
facienda edit 1 -t "New title" -m "New details"
```

### Database Location

By default, tasks are stored in `~/.facienda.db`. You can specify a custom database path:

```bash
facienda --db /path/to/custom.db list
```

## Project Structure

```
facienda/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   ├── add.go             # Add task command
│   ├── list.go            # List/view commands
│   ├── complete.go        # Complete/incomplete commands
│   └── edit.go            # Edit command
├── internal/              # Private application code
│   ├── todo/              # TODO domain logic
│   │   └── todo.go
│   └── storage/           # Data persistence
│       ├── storage.go     # Storage interface
│       ├── sqlite.go      # SQLite implementation
│       └── sqlite_test.go # Integration tests
├── main.go                # Application entry point
└── go.mod                 # Go module file
```

## Development

### Build

```bash
go build -o facienda .
```

### Run Tests

```bash
go test ./...
```

### Code Quality

```bash
go fmt ./...
go vet ./...
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Author

John Mirolha
