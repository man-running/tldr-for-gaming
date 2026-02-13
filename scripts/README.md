# Summary Generation Script

This script generates AI summaries of papers from the papers API and stores them in blob cache.

## Usage

```bash
# Run the script
./scripts/generate-summary

# Or with custom base URL
BASE_URL=https://your-domain.com ./scripts/generate-summary
```

## What it does

1. **Fetches papers data** from `/api/papers` endpoint (raw scraped feed)
2. **Generates AI summary** using OpenAI API
3. **Stores in blob cache** for fast retrieval
4. **Makes summary available** at `/api/tldr` (default endpoint)

## API Structure

| Endpoint | Purpose | Default Behavior |
|----------|---------|------------------|
| `/api/tldr` | **AI-generated summary** (RSS) | Primary endpoint for RSS readers |
| `/api/papers` | **Raw scraped papers** (RSS) | Raw feed data |

## Environment Variables

- `BASE_URL`: Base URL of your deployment (default: https://tldr.takara.ai)
- `OPENAI_API_KEY`: Required for summary generation
- `BLOB_READ_WRITE_TOKEN`: Required for blob storage

## Cron Job Example

Add to crontab for daily summary generation:

```bash
# Generate summary every day at 8 AM
0 8 * * * cd /path/to/tldr && ./scripts/generate-summary
```

## Data Flow

```
Papers API → Script → OpenAI → Blob Cache → TLDR API
    ↓         ↓        ↓        ↓         ↓
  scrape    fetch   summarize  store    serve
```

## RSS Reader Usage

**For RSS Readers:** Simply subscribe to `/api/tldr` - it serves AI-generated summaries by default.

**For Raw Data:** Use `/api/papers` for the original scraped feed data.
