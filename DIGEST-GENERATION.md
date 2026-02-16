# Digest Generation System

## Architecture Overview

This project uses a **background job** architecture for digest generation:

1. **Background Script** - Fetches articles, generates summaries, creates digest
2. **File Storage** - Stores everything as JSON files in `data/` directory
3. **API Routes** - Simply read from storage (fast, < 10ms)

## Directory Structure

```
data/
├── articles/
│   └── 2026-02-16.json          # Raw articles for each day
├── summaries/
│   └── newsapi-123456.json      # Individual article summaries
└── digests/
    └── 2026-02-16.json          # Generated digests
```

## Usage

### Generate a Digest (Manual)

Run this command to fetch articles, generate summaries, and create today's digest:

```bash
npm run generate-digest
```

This will:
1. Fetch gaming news from NewsAPI
2. Generate Claude AI summaries for top 5 articles (~60 seconds)
3. Create a narrative digest
4. Save everything to `data/` directory

### View the Digest

After generating, visit:
- **Digest page**: http://localhost:3001/digest
- **Individual summaries**: http://localhost:3001/summary/{article-id}

### Set Up Automated Generation (Cron)

To automatically generate digests daily, set up a cron job:

```bash
# Edit crontab
crontab -e

# Add this line to run daily at 6 AM:
0 6 * * * cd /path/to/tldr-for-gaming && npm run generate-digest >> logs/digest.log 2>&1
```

Or use a task scheduler like:
- **macOS**: launchd
- **Linux**: systemd timer
- **Cloud**: Vercel Cron, GitHub Actions, AWS EventBridge

### Environment Variables

Make sure these are set in `.env.local`:

```bash
CLAUDE_API_KEY=sk-ant-...
CLAUDE_MODEL=claude-opus-4-6
NEWSAPI_KEY=your-key-here
```

## API Endpoints

### GET /api/digest
- Returns today's digest (or latest available)
- Query param: `?date=2026-02-16` for specific date
- Fast response (~10ms) - reads from file

### GET /api/summary/[id]
- Returns summary for specific article
- Fast response (~5ms) - reads from file

## Storage Backend

Currently uses **file system** storage (`FileStorage` class), but designed to be easily upgradeable:

- **Current**: JSON files in `data/` directory
- **Future options**: SQLite, PostgreSQL, MongoDB

To migrate, just implement the `IStorage` interface in `lib/storage/storage-interface.ts`.

## Benefits

✅ **Fast API responses** - No Claude API calls during page load
✅ **Cost-effective** - Generate once, serve many times
✅ **Reliable** - No race conditions or timeouts
✅ **Scalable** - Can handle high traffic without hitting API rate limits
✅ **Debuggable** - Inspect JSON files directly
✅ **Flexible** - Easy to migrate to database later

## Troubleshooting

### No digest available
```bash
# Generate one manually:
npm run generate-digest
```

### Old digest showing
```bash
# Generate a fresh one:
npm run generate-digest
```

### Summary not found
The summary might not have been generated. Run the digest generator to create summaries for today's articles.

## Development

### File Locations

- **Storage interface**: `lib/storage/storage-interface.ts`
- **File storage**: `lib/storage/file-storage.ts`
- **Generator script**: `scripts/generate-digest.ts`
- **API routes**: `app/api/digest/route.ts`, `app/api/summary/[id]/route.ts`

### Testing

```bash
# Generate a test digest
npm run generate-digest

# Check the generated files
ls -la data/digests/
ls -la data/summaries/
ls -la data/articles/
```
