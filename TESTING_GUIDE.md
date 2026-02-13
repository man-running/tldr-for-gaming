# Phase 1 Testing Guide

## Overview

Phase 1 includes comprehensive test coverage for:
1. **Article Data Types** (`lib/article/types_test.go`) - 19 tests
2. **News Source Management** (`lib/feed/sources_test.go`) - 27 tests
3. **TypeScript Integration** (`lib/feed/sources.test.ts`) - 40+ test cases

**Total Coverage:** 86+ test cases across 3 files

---

## Running Tests

### Go Tests

#### Run All Go Tests
```bash
# From project root
go test ./lib/article/... ./lib/feed/...

# Or with verbose output
go test -v ./lib/article/... ./lib/feed/...

# With coverage report
go test -cover ./lib/article/... ./lib/feed/...

# Generate coverage profile
go test -coverprofile=coverage.out ./lib/article/... ./lib/feed/...
go tool cover -html=coverage.out
```

#### Run Specific Test Files
```bash
# Article types tests only
go test -v ./lib/article/types_test.go ./lib/article/types.go

# Sources tests only
go test -v ./lib/feed/sources_test.go ./lib/feed/sources.go
```

#### Run Specific Test
```bash
# Single test function
go test -run TestArticleDataCreation ./lib/article/...

# Tests matching pattern
go test -run TestRanking ./lib/article/...
```

### TypeScript/JavaScript Tests

#### Setup Jest/Vitest
```bash
# Install if not already installed
bun add -D jest @types/jest
# or
bun add -D vitest

# Create jest config (if needed)
bun create-jest-config
```

#### Run TypeScript Tests
```bash
# Run all tests
bun test

# Run tests in watch mode
bun test --watch

# Run with coverage
bun test --coverage

# Run specific test file
bun test lib/feed/sources.test.ts

# Run tests matching pattern
bun test --testNamePattern="getActiveSources"
```

#### With Jest
```bash
jest lib/feed/sources.test.ts
jest --coverage
jest --watch
```

#### With Vitest
```bash
vitest lib/feed/sources.test.ts
vitest --coverage
vitest --watch
```

---

## Test Structure

### Go Tests (`types_test.go`)

#### Test Categories:

**1. Creation & Initialization (4 tests)**
- `TestArticleDataCreation` - Create valid articles
- `TestArticleMetadataCreation` - Create metadata
- `TestRankedArticle` - Create ranked articles
- `TestDailyDigest` - Create daily digest

**2. Enumerations (2 tests)**
- `TestArticleCategories` - Validate category enum
- `TestErrorCodes` - Validate error codes

**3. Error Handling (1 test)**
- `TestApiError` - Test error responses

**4. Filtering & Configuration (3 tests)**
- `TestArticleFilter` - Test filter options
- `TestRankingCriteria` - Test ranking weights
- `TestRankingCriteriaCustomization` - Modify weights

**5. Data Validation (7 tests)**
- `TestArticleDataWithMetadata` - Complex metadata
- `TestArticleDataWithMultipleAuthors` - Multiple authors
- `TestArticleDataWithMultipleCategories` - Multiple categories
- `TestRankedArticleScoreValidation` - Score bounds (0-1)
- `TestDailyDigestDateFormat` - Date format validation

**Total: 19 tests**

### Go Tests (`sources_test.go`)

#### Test Categories:

**1. Initialization & Loading (4 tests)**
- `TestNewSourceManager` - Create manager
- `TestLoadDefaultSources` - Load 5 default sources
- `TestAddSource` - Add new source
- `TestAddSourceValidation` - Validation on add

**2. Retrieval & Listing (7 tests)**
- `TestGetSource` - Retrieve by ID
- `TestGetActiveSources` - Get active, sorted by priority
- `TestGetSourcesByCategory` - Filter by category
- `TestListSources` - List all sources
- `TestGetSourceCount` - Count total sources
- `TestGetActiveSourceCount` - Count active sources
- `TestLoadSourcesFromFile` - Load from JSON

**3. Modifications (3 tests)**
- `TestUpdateSource` - Update source fields
- `TestDisableSource` - Disable source
- `TestEnableSource` - Re-enable source

**4. Data Operations (2 tests)**
- `TestExportSources` - Export as JSON
- `TestSourceTimestamps` - Creation/update timestamps

**5. Validation (3 tests)**
- `TestValidate` - Validate configuration
- `TestValidateInvalidPriority` - Check priority bounds
- `TestValidateInvalidScrapingType` - Check scraping type

**6. Concurrency (1 test)**
- `TestConcurrentAccess` - Thread-safety with goroutines

**Total: 27 tests**

### TypeScript Tests (`sources.test.ts`)

#### Test Suites:

**1. DEFAULT_NEWS_SOURCES (5 tests)**
- Verify 5 sources exist
- Check required fields
- Validate priority values (1-10)
- Validate scraping types
- Confirm all active

**2. getSourceById (3 tests)**
- Return by valid ID
- Return undefined for invalid
- Check correct properties

**3. getActiveSources (3 tests)**
- Return active sources only
- Verify priority sort order
- Check first source is highest priority

**4. getSourcesByCategory (4 tests)**
- Return by category
- Empty for non-existent category
- Multiple sources sorted
- Filter by exact match

**5. getSourceMetadata (3 tests)**
- Return metadata for valid source
- Undefined for invalid
- Check correct fields only

**6. getCategories (4 tests)**
- Return unique categories
- Return sorted array
- No duplicates
- Only active categories

**7. isSourceActive (3 tests)**
- True for active sources
- False for non-existent
- Correct status for all

**8. getSourceCount (2 tests)**
- Return correct count
- Match default sources length

**9. getActiveSourceCount (3 tests)**
- Return active count
- Match filtered results
- Return 5 (all active)

**10. Default Sources Validation (4 tests)**
- iGamingBusiness config
- Gambling Insider config
- eGaming Review category
- Sports Betting category

**11. Edge Cases (4 tests)**
- Case-sensitive searches
- Empty string handling
- Null/undefined safety

**12. Data Integrity (3 tests)**
- Immutability of defaults
- Separate array instances
- Prevent modifications

**13. Performance (1 test)**
- Quick retrieval (<100ms)

**Total: 40+ test cases**

---

## Coverage Goals

### Target Coverage
- **Go Code:** 90%+ coverage
- **TypeScript Code:** 85%+ coverage

### Generate Coverage Reports

#### Go Coverage
```bash
go test -coverprofile=coverage.out ./lib/article/... ./lib/feed/...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

#### TypeScript Coverage
```bash
bun test --coverage

# Or with Jest
jest --coverage

# Or with Vitest
vitest --coverage
```

---

## Test Scenarios

### Scenario 1: Loading Sources
```bash
go test -run "TestLoadDefaultSources" -v ./lib/feed/...
```
**Expected:** Loads 5 sources successfully

### Scenario 2: Priority Sorting
```bash
go test -run "TestGetActiveSources" -v ./lib/feed/...
```
**Expected:** Sources sorted by priority descending (10, 9, 8, 7, 7)

### Scenario 3: Category Filtering
```bash
bun test --testNamePattern="getSourcesByCategory"
```
**Expected:** Returns only sources in specified category

### Scenario 4: Concurrent Access
```bash
go test -run "TestConcurrentAccess" -v ./lib/feed/...
```
**Expected:** All goroutines complete without race conditions

### Scenario 5: Validation
```bash
go test -run "Validate" -v ./lib/feed/...
```
**Expected:** Catches invalid configurations

---

## Common Issues & Solutions

### Issue: "fatal error: runtime: concurrent map iteration"
**Solution:** This indicates a race condition. Run with `-race` flag:
```bash
go test -race ./lib/feed/...
```

### Issue: Jest/Vitest not found
**Solution:** Install test framework:
```bash
bun add -D jest @types/jest
# or
bun add -D vitest
```

### Issue: Go tests fail with import errors
**Solution:** Ensure you're in the correct directory:
```bash
cd /Volumes/Extreme/dev/tldr-for-gaming
go test ./lib/article/... ./lib/feed/...
```

### Issue: TypeScript compilation errors
**Solution:** Verify TypeScript setup:
```bash
bun check  # Check TypeScript compilation
```

---

## Adding New Tests

### Go Test Template
```go
func TestNewFeature(t *testing.T) {
    // Arrange
    manager := NewSourceManager()

    // Act
    result := manager.SomeMethod()

    // Assert
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### TypeScript Test Template
```typescript
test("should do something", () => {
    // Arrange
    const input = getSourceById("igamingbusiness");

    // Act
    const result = someFunction(input);

    // Assert
    expect(result).toBe(expected);
});
```

---

## Continuous Integration

### GitHub Actions Configuration (Optional)

Create `.github/workflows/test.yml`:
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

---

## Test Checklist

Before moving to Phase 2:

- [ ] All Go tests pass: `go test -v ./lib/article/... ./lib/feed/...`
- [ ] No race conditions: `go test -race ./lib/article/... ./lib/feed/...`
- [ ] Go coverage > 85%: `go test -cover ./lib/article/... ./lib/feed/...`
- [ ] All TypeScript tests pass: `bun test`
- [ ] TypeScript coverage > 80%: `bun test --coverage`
- [ ] No linting errors: `bun run lint` (if configured)
- [ ] All type checks pass: `bun check`
- [ ] Documentation updated: `TESTING_GUIDE.md` ✅

---

## Performance Benchmarks

### Go Benchmarks (Optional)

Create `lib/feed/sources_bench_test.go`:
```go
func BenchmarkGetActiveSources(b *testing.B) {
    manager := NewSourceManager()
    manager.LoadDefaultSources()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        manager.GetActiveSources()
    }
}
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./lib/feed/...
```

---

## Next Steps

After Phase 1 testing is complete:

1. ✅ Run all tests and verify they pass
2. ✅ Generate coverage reports
3. ✅ Document test results
4. ✅ Commit tests to git
5. **→ Proceed to Phase 2: Feed Integration**

---

## Resources

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [Vitest Documentation](https://vitest.dev/)
- [Go Race Detector](https://golang.org/doc/articles/race_detector)

---

**Test Coverage Status:** 86+ tests ready for execution
**Estimated Test Duration:**
- Go tests: <5 seconds
- TypeScript tests: <10 seconds
- Total: <15 seconds

