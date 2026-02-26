# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

yytools is a Go utilities library providing algorithms, data structures, and common tools for daily development. The project uses a flat package structure, with a focus on performance and correctness.

## Architecture

### Core Structure
- `pkg/`: Core functionality organized by domain
  - `algorithms/`: Algorithm implementations (sorting, math utilities, binary search)
    - `binary_search/`: Binary search implementation
    - `mathutils/`: GCD, probability distributions, common math
      - `bits/`: Bit operation utilities
      - `overflow/`: Numeric overflow check utilities
      - `probability_distribution/`: Probability distribution utilities
      - `random/`: Random number generation
    - `sort/`: Multiple sorting algorithm implementations
  - `common/`: Common utilities
    - `assert/`: Runtime assertion framework that can be toggled on/off
    - `base/`: Type definitions and constraints for generics
    - `os/`: OS utility wrappers
    - `safeexec/`: panic-safe function execution wrappers (Safe, SafeCall, SafeExecWithError, etc.)
    - `timeutils/`: Time utility functions
  - `concurrency/`: Concurrency control utilities
    - `unbounded_channel/`: Unbounded channel implementation with multiple variants
  - `datastructures/`: Data structure implementations (heap, queue, stack, sorted_set)

- `cmd/demo/`: Demo application with CLI and HTTP visualization server

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
go test ./pkg/algorithms/mathutils/random/

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
- All code uses the custom assertion framework in `pkg/common/assert/`
- Assertions can be enabled/disabled globally via `assert.SetAssert(bool)`
- Use `assert.Assert(condition, message...)` for conditional assertions
- Use `assert.AssertFast(condition)` for simple boolean checks

### Code Conventions
- Generic implementations use type constraints from `pkg/common/base/`
- Chinese comments are used throughout the codebase
- Apache 2.0 license header in all files
- Flat package structure: no `concrete/generic` intermediate directories except where both truly exist

### Performance Considerations
- Sorting algorithms include performance comparison with Go's standard library
- Use `cmd/demo/graph.go` for performance visualization at `http://localhost:8081`
- Counting sort and quick sort implementations are optimized for different data ranges