# Deck Package Tests

This package has two distinct types of tests:

## 1. Unit Tests (`unit-tests/`)

**Purpose**: Test our deck package implementation with focused, small test cases.

**Structure**:
```
pkg/deck/unit-tests/
├── input/           # Test DSH files we created
├── expected/        # Expected output files (SVG, PNG, PDF, XML)
├── output/          # Generated outputs for comparison
└── *.dsh           # Individual test files
```

**Usage**:
- Add focused test cases here for specific deck functionality
- Used by golden test runner for regression testing
- Small, targeted examples to test edge cases

## 2. Repo Tests (`repo-tests/`)

**Purpose**: Comprehensive validation against 525+ real-world examples from upstream repositories.

**Structure**:
```
pkg/deck/repo-tests/
├── [upstream repo contents from decksh and deck repositories]
├── dubois-data-portraits/    # Historical data visualization examples
├── [many other example directories...]
└── [525+ DSH files total]
```

**Usage**:
- Cloned from upstream repositories: 
  - `https://github.com/ajstarks/decksh.git`
  - `https://github.com/ajstarks/deck.git`
- Used by builder to compile tools from source
- Provides massive test coverage against real-world examples
- Automatically updated when repos are refreshed

## Running Tests

### Unit Tests
```bash
# Run golden tests against our unit test cases
go run . deck test

# Watch unit test directory for changes
go run . tools deck watch pkg/deck/unit-tests
```

### Repo Tests
```bash
# Build tools from repo tests (source code)
go run . deck build install

# The repo tests are used as source for building the deck tools
# and provide comprehensive examples for validation
```

## Adding New Tests

### For Unit Tests
1. Add `.dsh` file to `unit-tests/input/`
2. Generate expected outputs in `unit-tests/expected/`
3. Run golden test runner to validate

### For Repo Tests
- These are managed automatically via git submodules/cloning
- Updated when upstream repositories are refreshed
- No manual additions needed

## Test Philosophy

- **Unit tests**: Focus on testing our implementation
- **Repo tests**: Validate against real-world usage patterns
- Both provide complementary coverage for robust testing