# Integration Testing Guide

## Real Website Fetching Tests

The integration tests actually fetch articles from the real iGaming news websites to verify the fetcher works end-to-end with real data.

---

## Test Files

### `lib/feed/fetcher_integration_test.go`

Integration tests that make real HTTP requests to iGaming news sources.

**Tests Included:**
1. `TestFetchFromIgamingBusiness` - Fetch from iGamingBusiness.com
2. `TestFetchFromGamblingInsider` - Fetch from GamblingInsider.com
3. `TestFetchFromMultipleSources` - Fetch from all configured sources
4. `TestFetchAndCacheIntegration` - Full pipeline (fetch → cache → query)
5. `TestArticleQuality` - Verify article data quality
6. `TestRealDateParsing` - Test date parsing with real data
7. `BenchmarkFetchFromIgamingBusiness` - Performance benchmarking

---

## Running Integration Tests

### Run All Integration Tests
```bash
go test -v ./lib/feed/fetcher_integration_test.go ./lib/feed/fetcher.go
```

### Run Specific Test
```bash
go test -run TestFetchFromIgamingBusiness -v ./lib/feed/...
```

### Run with Custom Timeout
```bash
# 60 second timeout (recommended for slow networks)
go test -timeout 60s -v -run "Integration" ./lib/feed/...

# Or for specific test
go test -timeout 30s -v -run TestFetchFromMultipleSources ./lib/feed/...
```

### Skip Integration Tests (Quick tests only)
```bash
go test -short -v ./lib/feed/...
```

### Run Benchmarks
```bash
go test -bench BenchmarkFetch -benchtime=1x ./lib/feed/...
```

---

## What These Tests Do

### 1. TestFetchFromIgamingBusiness
**Fetches from:** iGamingBusiness.com RSS feed

**Verifies:**
- Successfully connects to real website
- Parses RSS feed correctly
- Extracts article data
- Articles have required fields (ID, Title, URL, etc)

**Example Output:**
```
Successfully fetched 15 articles from iGamingBusiness
```

### 2. TestFetchFromGamblingInsider
**Fetches from:** GamblingInsider.com RSS feed

**Verifies:**
- Different source configuration works
- Consistent parsing across sources
- Source metadata tracking

### 3. TestFetchFromMultipleSources
**Fetches from:** All 5 configured sources in parallel

**Verifies:**
- Batch fetching works
- Handles partial failures gracefully
- Aggregates articles from multiple sources
- Reports statistics per source

**Example Output:**
```
Successfully fetched 48 articles from 3 sources
  iGamingBusiness: 15 articles
  Gambling Insider: 18 articles
  eGaming Review: 15 articles
```

### 4. TestFetchAndCacheIntegration
**Full pipeline test:**
1. Fetch articles from real source
2. Cache articles in memory
3. Query cache by category
4. Verify caching works with real data

**Verifies:**
- End-to-end pipeline works
- Caching integrates with fetcher
- Filtering works with real article data
- Cache statistics are accurate

### 5. TestArticleQuality
**Verifies article data quality:**
- All required fields present
- No empty strings
- Valid timestamps
- Proper data types

**Checks:**
- ✅ ID (not empty)
- ✅ Title (not empty)
- ✅ URL (valid URL)
- ✅ SourceName (correct)
- ✅ PublishedDate (RFC3339 format)
- ✅ Categories (at least one)
- ✅ CreatedAt (valid timestamp)

### 6. TestRealDateParsing
**Verifies date parsing with real dates:**
- Dates are correctly parsed
- Articles are recent (within 7 days)
- No future-dated articles

**Warnings:**
- Logs if articles are older than 7 days
- Logs if articles have future dates

### 7. BenchmarkFetchFromIgamingBusiness
**Performance benchmarking:**
- Measures time to fetch from single source
- Reports ops/sec and time per operation
- Identifies performance bottlenecks

---

## Example: Running Tests

### Quick Test (Unit Tests Only)
```bash
$ go test -short -v ./lib/feed/...
=== RUN   TestNewArticleFetcher
--- PASS: TestNewArticleFetcher (0.00s)
...
ok      command-line-arguments  0.123s
```

### Full Test (Including Integration)
```bash
$ go test -v -timeout 60s ./lib/feed/fetcher_integration_test.go ./lib/feed/fetcher.go
=== RUN   TestFetchFromIgamingBusiness
Successfully fetched 15 articles from iGamingBusiness
--- PASS: TestFetchFromIgamingBusiness (3.45s)

=== RUN   TestFetchAndCacheIntegration
Cache stats - Total: 15, Valid: 15, Utilization: 1.5%
Found 8 articles in category Business
Successfully fetched, cached, and queried 15 articles
--- PASS: TestFetchAndCacheIntegration (3.67s)

ok      command-line-arguments  7.123s
```

---

## Handling Network Issues

### Test Skips (Not Failures)
The integration tests **skip** (not fail) when:
- Network is unavailable
- Websites are down
- No articles returned

This prevents false failures due to network/server issues.

### If Tests Fail

**Timeout errors:**
```bash
# Increase timeout
go test -timeout 120s ./lib/feed/...
```

**Connection refused:**
- Check internet connection
- Check if websites are accessible
- Try again later if sites are down

**Parse errors:**
- Website HTML/RSS structure may have changed
- May need to update parser for that source

---

## Expected Test Results

### Success Example
```
Testing iGamingBusiness.com...
Successfully fetched 12 articles
Article 1: "New Regulations in UK" - https://example.com/article-1
Article 2: "Casino Expansion News" - https://example.com/article-2
...

Testing all sources...
Successfully fetched 48 articles from 3 sources
Cache quality: 100% valid articles
```

### Skip Example (Network Down)
```
TestFetchFromIgamingBusiness: SKIP
Reason: No articles fetched (site may be down or empty)
```

### Timeout Example
```
TestFetchFromMultipleSources: TIMEOUT
Reason: Test exceeded 30s timeout
Solution: Use -timeout 60s flag
```

---

## Continuous Integration

### GitHub Actions Setup

Add to `.github/workflows/tests.yml`:

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run integration tests
        run: |
          go test -v -timeout 120s ./lib/feed/fetcher_integration_test.go ./lib/feed/fetcher.go
        continue-on-error: true  # Don't fail on network issues
```

---

## Development Workflow

### Before Committing
```bash
# Run unit tests (fast, no network)
go test -short -v ./lib/feed/...

# Run integration tests (slower, needs network)
go test -timeout 60s -v ./lib/feed/fetcher_integration_test.go ./lib/feed/fetcher.go
```

### During Development
```bash
# Quick feedback loop
go test -short -v ./lib/feed/... && echo "Unit tests passed!"

# Full validation before push
go test -v ./lib/feed/... && go test -timeout 60s -v ./lib/feed/fetcher_integration_test.go ./lib/feed/fetcher.go
```

### In CI/CD
```bash
# Run with timeouts and error handling
go test -timeout 120s -v ./lib/feed/... || true
```

---

## Benchmarking Performance

### Run Benchmarks
```bash
# Single iteration
go test -bench BenchmarkFetch -benchtime=1x ./lib/feed/...

# Multiple iterations
go test -bench BenchmarkFetch -count=5 ./lib/feed/...

# With memory stats
go test -bench BenchmarkFetch -benchmem ./lib/feed/...
```

### Example Output
```
BenchmarkFetchFromIgamingBusiness-8    1        3450000000 ns/op        1.2 MB/op
```

This means:
- 3.45 seconds per fetch
- ~1.2 MB memory allocation per fetch

---

## Monitoring Article Feed Health

### Test Results Interpretation

**Good Signs:**
- ✅ Consistent article count (10+ per source)
- ✅ Recent publication dates (within 7 days)
- ✅ All articles have required fields
- ✅ Fast fetch times (< 5s per source)

**Warning Signs:**
- ⚠️ Declining article count
- ⚠️ No recent articles
- ⚠️ Slow fetch times (> 10s)
- ⚠️ Frequent timeouts

**Troubleshooting:**
- Check website status manually
- Verify RSS feed URL is still valid
- Check if HTML structure changed
- Review rate limiting headers

---

## Testing Each Source Individually

```bash
# Test iGamingBusiness
go test -run TestFetchFromIgamingBusiness -v -timeout 30s ./lib/feed/...

# Test Gambling Insider
go test -run TestFetchFromGamblingInsider -v -timeout 30s ./lib/feed/...

# Test all sources
go test -run TestFetchFromMultipleSources -v -timeout 60s ./lib/feed/...
```

---

## Common Issues & Solutions

### Issue: "context deadline exceeded"
**Cause:** Network is slow or site is slow
**Solution:** Increase timeout
```bash
go test -timeout 120s -v ./lib/feed/...
```

### Issue: "no articles fetched"
**Cause:** RSS feed might be empty or site is down
**Solution:** Check site manually, test will skip not fail

### Issue: "invalid date format"
**Cause:** RSS feed uses unusual date format
**Solution:** May need to add format support to `parsePublishDate()`

### Issue: "connection refused"
**Cause:** Internet connection issue
**Solution:** Check network connectivity
```bash
ping google.com  # Test internet
curl https://www.igamingbusiness.com/feed/  # Test website
```

---

## Next Steps

### Phase 3 Integration Testing
When summarization is added:
```bash
# Test fetch + summarize pipeline
go test -run "Integration.*Summarize" -v ./lib/feed/...
```

### Phase 4 Integration Testing
When database is added:
```bash
# Test fetch + cache + database pipeline
go test -run "Integration.*Database" -v ./lib/feed/...
```

---

## Resources

- [Testing Go code](https://golang.org/doc/effective_go#testing)
- [HTTP testing in Go](https://golang.org/pkg/net/http/httptest/)
- [Context package](https://golang.org/pkg/context/)

---

**Last Updated:** February 13, 2026
**Test Coverage:** Real website fetching with 7 integration tests
