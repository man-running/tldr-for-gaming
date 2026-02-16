# Vercel Deployment Guide - Blob Storage Edition

## Overview

✅ **This application is now ready for Vercel deployment!** It uses **Vercel Blob Storage** for persistent data and **Vercel Cron Jobs** for automated digest generation.

---

## Architecture

### Dual Storage System

The app uses an **environment-aware storage factory** that automatically switches between:

- **Local Development**: File-based storage (`data/` directory)
- **Vercel Production**: Blob Storage (S3-like object storage)

### Storage Structure
```
Blob Storage (Production):       File Storage (Local):
├── articles/2026-02-16.json     ├── data/articles/2026-02-16.json
├── summaries/newsapi-123.json   ├── data/summaries/newsapi-123.json
└── digests/2026-02-16.json      └── data/digests/2026-02-16.json
```

### Article Sources

The digest fetches from **5 sources in parallel**:
1. **NewsAPI** - Generic gambling/gaming news
2. **SBC News** (RSS) - Sports Betting Community
3. **iGaming Business** (RSS) - Industry news
4. **Calvin Ayre** (RSS) - Gaming industry news
5. **EGR** (RSS) - eGaming Review

### Automated Digest Generation
- **Production**: Vercel Cron runs daily at 6:00 AM UTC via `/api/cron/generate-digest`
- **Local**: Manual execution with `npm run generate-digest`

### API Routes
- `/api/digest` - Returns latest/specific digest
- `/api/summary/[id]` - Returns individual article summary
- `/api/cron/generate-digest` - Protected cron endpoint (requires `CRON_SECRET`)

---

## Why Blob Storage?

**Blob Storage** was chosen over Postgres because:
- ✅ File-like storage (matches existing architecture)
- ✅ No schema migrations needed
- ✅ **Cheapest option** (~$0.03/month vs $0.30/month for Postgres)
- ✅ Simple to implement (same JSON structure)
- ✅ Perfect for document/object storage
- ✅ No complex queries needed

---

## What's Already Implemented

✅ **Blob Storage Class** (`lib/storage/blob-storage.ts`)
- Implements same `IStorage` interface as FileStorage
- Uses Vercel Blob's `put()`, `list()` APIs
- Maintains same path structure

✅ **Storage Factory** (`lib/storage/index.ts`)
- Auto-detects environment (`VERCEL` or `BLOB_READ_WRITE_TOKEN`)
- Returns BlobStorage on Vercel, FileStorage locally

✅ **Digest Generator Module** (`lib/feed/digest-generator.ts`)
- Extracted from script into reusable library
- Accepts storage interface as parameter
- Used by both local script and cron job

✅ **Cron API Endpoint** (`app/api/cron/generate-digest/route.ts`)
- Protected with `CRON_SECRET` authentication
- 5-minute timeout (digest takes ~70-75 seconds)
- Returns success/error status

✅ **Cron Configuration** (`vercel.json`)
- Schedules daily execution at 6:00 AM UTC
- Can be customized (every 4 hours, weekly, etc.)

✅ **Updated API Routes**
- Both `/api/digest` and `/api/summary/[id]` use storage factory
- Work automatically in both environments

---

## Deployment Steps

### 1. Commit Your Changes

```bash
git add .
git commit -m "Add Vercel Blob storage and automated cron digest generation"
git push origin main
```

### 2. Connect to Vercel

1. Go to [vercel.com](https://vercel.com)
2. Click **"New Project"**
3. Import your Git repository (GitHub/GitLab/Bitbucket)
4. Vercel auto-detects Next.js configuration
5. Click **"Deploy"** (don't worry about env vars yet - we'll add them next)

### 3. Enable Vercel Blob Storage

**This is the key step for persistent storage:**

1. Go to your deployed project in Vercel Dashboard
2. Navigate to **Storage** tab
3. Click **"Create Database"** or **"Create Store"**
4. Select **"Blob"** ← **IMPORTANT: Choose Blob, NOT Postgres!**
5. Choose a store name (e.g., "tldr-gaming-blobs")
6. Click **"Create"**

✅ Vercel automatically creates `BLOB_READ_WRITE_TOKEN` environment variable

### 4. Configure Environment Variables

Go to **Settings → Environment Variables** and add:

| Variable | Value | Description |
|----------|-------|-------------|
| **`CLAUDE_API_KEY`** | `sk-ant-...` | Your Anthropic API key for summaries |
| **`NEWSAPI_KEY`** | `your-key` | Your NewsAPI key for fetching articles |
| **`CRON_SECRET`** | Generate random | Protects cron endpoint from unauthorized access |
| **`CLAUDE_MODEL`** | `claude-opus-4-6` | (Optional) Model to use, has default |
| **`BLOB_READ_WRITE_TOKEN`** | *Auto-created* | Automatically set by Vercel when you create Blob store |

**Generate CRON_SECRET:**
```bash
openssl rand -base64 32
# or
node -e "console.log(require('crypto').randomBytes(32).toString('base64'))"
```

**Important:** Set all variables for **Production**, **Preview**, and **Development** environments.

### 5. Redeploy

After adding environment variables, trigger a new deployment:
- Go to **Deployments** tab
- Click "..." on latest deployment → **"Redeploy"**
- Or push a new commit to trigger auto-deployment

### 6. Generate Initial Digest

Manually trigger the first digest to populate Blob Storage:

```bash
curl -X GET \
  -H "Authorization: Bearer YOUR_CRON_SECRET" \
  https://your-app.vercel.app/api/cron/generate-digest
```

**Expected Response:**
```json
{
  "success": true,
  "date": "2026-02-16",
  "articleCount": 5,
  "message": "Digest generated successfully"
}
```

### 7. Verify Everything Works

#### Check Blob Storage
1. Vercel Dashboard → Storage → Blob
2. Should see files with prefixes:
   - `articles/`
   - `summaries/`
   - `digests/`

#### Test API Endpoints
```bash
# Get latest digest
curl https://your-app.vercel.app/api/digest

# Get specific summary
curl https://your-app.vercel.app/api/summary/newsapi-123456
```

#### Verify Cron Job
1. Vercel Dashboard → Deployments → Cron Jobs tab
2. Should see `/api/cron/generate-digest` scheduled
3. Wait until 6:00 AM UTC or manually trigger again

---

## Local Development

The storage factory automatically uses **FileStorage** locally - no changes needed!

```bash
# Uses local file storage (data/ directory)
npm run generate-digest

# Start dev server (reads from local data/)
npm run dev

# Visit http://localhost:3001/digest
```

Your local `data/` directory continues to work exactly as before.

---

## Customizing the Cron Schedule

Edit `vercel.json` to change when digests are generated:

```json
{
  "crons": [
    {
      "path": "/api/cron/generate-digest",
      "schedule": "0 6 * * *"  // ← Change this
    }
  ]
}
```

**Common Schedules:**
- `"0 6 * * *"` - Daily at 6:00 AM UTC
- `"0 */4 * * *"` - Every 4 hours
- `"0 12 * * 1"` - Every Monday at 12:00 PM UTC
- `"0 0 * * *"` - Midnight UTC daily
- `"0 9 * * 1-5"` - Weekdays at 9:00 AM UTC

After changing, push to trigger redeployment.

---

## Cost Estimates

### Vercel Blob Storage Pricing

| Resource | Free Tier | Cost After |
|----------|-----------|------------|
| **Storage** | - | $0.15/GB/month |
| **Reads** | 500K/month | $0.15/million |
| **Writes** | 500K/month | $0.15/million |

### Estimated Monthly Usage

For daily digests:
- **Storage**: ~500 KB/day × 365 days = ~180 MB/year = **$0.03/month**
- **Writes**: 1 digest + 15 summaries + 1 articles = ~17 writes/day = **Free tier**
- **Reads**: ~1000 page views/day = 1000 reads/day = **Free tier**

**Total Cost: ~$0.03/month** (essentially free!)

### Comparison

| Storage Type | Cost/Month | Setup Complexity |
|--------------|------------|------------------|
| **Blob Storage** | ~$0.03 | ⭐ Simple |
| Vercel Postgres | ~$0.30+ | ⭐⭐ Medium |
| Vercel KV | ~$0.20+ | ⭐ Simple |

---

## Troubleshooting

### "No digest available" Error

**Problem**: API returns 404 when accessing `/api/digest`

**Solution**:
1. Check Blob Storage has files (Vercel Dashboard → Storage → Blob)
2. Manually trigger cron to generate first digest:
   ```bash
   curl -H "Authorization: Bearer YOUR_CRON_SECRET" \
     https://your-app.vercel.app/api/cron/generate-digest
   ```

### Cron Returns 401 Unauthorized

**Problem**: Cron endpoint returns "Unauthorized"

**Solution**:
1. Verify `CRON_SECRET` is set in Vercel environment variables
2. Check you're using correct secret in Authorization header
3. Format must be: `Bearer YOUR_SECRET` (note the space)

### Function Timeout

**Problem**: Digest generation times out

**Solution**: The cron endpoint already has `maxDuration: 300` (5 minutes). If it still times out:

1. Check if Claude API is responding slowly
2. Consider reducing `articlesToSummarize` from 15 to 10 in `digest-generator.ts`
3. Or implement parallel summary generation (currently serial)

### BLOB_READ_WRITE_TOKEN Not Set

**Problem**: Error about missing BLOB_READ_WRITE_TOKEN

**Solution**:
1. Ensure you created a Blob store (Step 3 above)
2. Vercel automatically sets this token - you don't need to add it manually
3. Try redeploying after creating the Blob store

### Local Development Not Working

**Problem**: `npm run generate-digest` fails

**Solution**:
1. Ensure `.env.local` has `CLAUDE_API_KEY` and `NEWSAPI_KEY`
2. Check `data/` directory exists (should be created automatically)
3. Local development uses FileStorage, not Blob - no BLOB_READ_WRITE_TOKEN needed

---

## Monitoring & Logs

### View Cron Execution Logs

1. Vercel Dashboard → Deployments
2. Click on a deployment → Functions tab
3. Find `/api/cron/generate-digest`
4. Click to see execution logs

### Check Generated Digests

```bash
# List all blobs
curl https://[your-blob-store-url].blob.vercel-storage.com/?prefix=digests/

# View specific digest (in browser)
https://your-app.vercel.app/api/digest?date=2026-02-16
```

---

## Security Best Practices

### Protecting the Cron Endpoint

✅ **Already Implemented:**
- `CRON_SECRET` authentication required
- Only Vercel's cron system knows the secret
- API endpoint validates Bearer token

### Environment Variables

✅ **Best Practices:**
- Never commit `.env.local` to git (already in `.gitignore`)
- Rotate `CRON_SECRET` if accidentally exposed
- Use different secrets for Production/Preview/Development

### Blob Access

✅ **Current Setup:**
- Blobs are set to `access: 'public'` for API reads
- No authentication needed to read digest/summary JSON
- This is intentional - it's public content

If you need private blobs, change `access: 'public'` to `access: 'private'` in `blob-storage.ts`.

---

## Next Steps

### After Successful Deployment

1. **Monitor First Cron Run**: Wait until 6:00 AM UTC and check logs
2. **Test Digest Page**: Visit `https://your-app.vercel.app/digest`
3. **Set Up Analytics**: Consider adding Vercel Analytics
4. **Custom Domain**: Add custom domain in Vercel settings
5. **Preview Deployments**: Test changes in preview before production

### Future Enhancements

- **Parallel Summary Generation**: Speed up digest creation (60s → 5s)
- **Edge Caching**: Serve digests from CDN
- **Email Notifications**: Send digest via email when generated
- **RSS Feed**: Generate RSS feed from digests
- **Multiple Timezones**: Generate digests for different regions

---

## Resources

- [Vercel Blob Documentation](https://vercel.com/docs/storage/vercel-blob)
- [Vercel Cron Jobs Documentation](https://vercel.com/docs/cron-jobs)
- [Next.js Deployment Documentation](https://nextjs.org/docs/deployment)
- [Environment Variables](https://vercel.com/docs/concepts/projects/environment-variables)

---

## Support

If you encounter issues:

1. Check Vercel function logs (Deployments → Functions)
2. Review this guide's Troubleshooting section
3. Verify all environment variables are set
4. Ensure Blob store was created (not Postgres!)

**Common mistake**: Creating Postgres instead of Blob store. If you did this, delete the Postgres store and create a Blob store instead.
