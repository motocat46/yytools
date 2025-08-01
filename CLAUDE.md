# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

yytools is a Go utilities library providing algorithms, data structures, and common tools for daily development. The project is organized into concrete (non-generic) and generic implementations, with a focus on performance and correctness.

## Architecture

### Core Structure
- `pkg/`: Core functionality organized by domain
  - `algorithms/concrete/`: Non-generic algorithm implementations (sorting, math utilities, binary search)
  - `common/concrete/`: Common utilities (assertions, base types, OS utilities, time utilities)
  - `datastructures/concrete/`: Data structure implementations
  - `concurrency/concrete/`: Concurrency control utilities
  - `gameutils/concrete/`: Game-specific utilities
  - `*/generic/`: Generic versions of corresponding concrete implementations

- `datastructure/`: Legacy data structures (heap, queue, stack, sorted_set) - still in use
- `template/`: Code templates and reference implementations
- `examples/`: Usage examples
- `tests/`: Integration and benchmark tests

### Key Components
- **Assertion System**: `pkg/common/concrete/assert/` - Runtime assertion framework that can be toggled on/off
- **Base Types**: `pkg/common/concrete/base/` - Type definitions and constraints for generics
- **Math Utilities**: `pkg/algorithms/concrete/mathutils/` - GCD, probability distributions, random number generation
- **Sorting Algorithms**: `pkg/algorithms/concrete/sort/` - Multiple sorting implementations with performance testing
- **Data Structures**: Both legacy (`datastructure/`) and new (`pkg/datastructures/`) implementations

## Common Commands

### Building and Running
```bash
# Build the project
go build

# Run specific functionality tests
go run . <command> <iterations>

# Available commands (get help)
go run . help
```

### Testing Individual Components
```bash
# Test heap operations (5 iterations)
go run . heap 5

# Test sorting algorithms (10 iterations)  
go run . sort 10

# Test all components (3 iterations)
go run . all 3

# Start performance visualization server
go run . http
```

### Standard Go Commands
```bash
# Run all Go tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test file
go test ./pkg/algorithms/concrete/mathutils/random/

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

## Development Notes

### Testing Strategy
- Unit tests follow `*_test.go` naming convention
- Custom test functions use `Test*` naming pattern (e.g., `HeapTest`, `SortTest`)
- Test iterations are configurable via command line arguments
- Performance testing includes visualization via go-echarts

### Assertion System
- All code uses the custom assertion framework in `pkg/common/concrete/assert/`
- Assertions can be enabled/disabled globally via `assert.SetAssert(bool)`
- Use `assert.Assert(condition, message...)` for conditional assertions
- Use `assert.AssertFast(condition)` for simple boolean checks

### Code Conventions
- Generic implementations use type constraints from `pkg/common/concrete/base/`
- Chinese comments are used throughout the codebase
- Apache 2.0 license header in all files
- Consistent naming: concrete vs generic package separation

### Performance Considerations
- Sorting algorithms include performance comparison with Go's standard library
- Use `graph.go` for performance visualization at `http://localhost:8081`
- Counting sort and quick sort implementations are optimized for different data ranges