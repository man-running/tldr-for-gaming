# Vercel Deployment Guide

## Overview

This Next.js application can be deployed to Vercel, but requires modifications to work with Vercel's serverless architecture. The main issue is that Vercel functions have **ephemeral file systems** - files written during runtime don't persist.

---

## Current Architecture (Local Development)

### File-Based Storage
```
data/
├── articles/2026-02-16.json      # Raw articles
├── summaries/newsapi-123.json    # Individual summaries
└── digests/2026-02-16.json       # Generated digests
```

### Background Job
- Manual execution: `npm run generate-digest`
- Fetches articles, generates summaries, creates digest
- Saves everything to local `data/` directory

### API Routes
- `/api/digest` - Reads from `data/digests/`
- `/api/summary/[id]` - Reads from `data/summaries/`

---

## Why Current Architecture Won't Work on Vercel

1. **Ephemeral File System**: Vercel serverless functions can write files during execution, but they disappear after the function finishes or between deployments.

2. **No Persistent Disk**: The `data/` directory would be lost, making all generated digests and summaries inaccessible.

3. **No Background Processes**: Can't run `npm run generate-digest` as a traditional cron job.

---

## Required Changes for Vercel Deployment

### 1. Migrate from File Storage to Database

Replace `FileStorage` implementation with a database-backed storage system.

#### Option A: Vercel Postgres (Recommended)

**Pros:**
- Native Vercel integration
- SQL database (familiar, powerful)
- Easy to set up
- Free tier available

**Setup:**
1. Add Vercel Postgres to your project via Vercel dashboard
2. Install dependencies:
   ```bash
   npm install @vercel/postgres
   ```

3. Create database schema:
   ```sql
   CREATE TABLE articles (
     id VARCHAR(255) PRIMARY KEY,
     title TEXT NOT NULL,
     summary TEXT,
     original_summary TEXT,
     url TEXT NOT NULL,
     source_name VARCHAR(255),
     source_id VARCHAR(255),
     published_date TIMESTAMP,
     image_url TEXT,
     categories JSONB,
     created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE summaries (
     id VARCHAR(255) PRIMARY KEY,
     title TEXT NOT NULL,
     summary TEXT NOT NULL,
     url TEXT NOT NULL,
     source_name VARCHAR(255),
     published_date TIMESTAMP,
     generated_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE digests (
     date DATE PRIMARY KEY,
     articles JSONB NOT NULL,
     summary TEXT NOT NULL,
     created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE INDEX idx_articles_published ON articles(published_date DESC);
   CREATE INDEX idx_summaries_generated ON summaries(generated_at DESC);
   ```

4. Create new storage implementation (`lib/storage/postgres-storage.ts`):
   ```typescript
   import { sql } from '@vercel/postgres';
   import type { IStorage, Article, StoredSummary, DailyDigest } from './storage-interface';

   export class PostgresStorage implements IStorage {
     async initialize(): Promise<void> {
       // Tables created via SQL above
     }

     async saveArticles(date: string, articles: Article[]): Promise<void> {
       for (const article of articles) {
         await sql`
           INSERT INTO articles (id, title, summary, original_summary, url, source_name, source_id, published_date, image_url, categories)
           VALUES (${article.id}, ${article.title}, ${article.summary}, ${article.originalSummary}, ${article.url}, ${article.sourceName}, ${article.sourceId}, ${article.publishedDate}, ${article.imageUrl || null}, ${JSON.stringify(article.categories || [])})
           ON CONFLICT (id) DO UPDATE SET
             title = EXCLUDED.title,
             summary = EXCLUDED.summary,
             url = EXCLUDED.url
         `;
       }
     }

     async getArticles(date: string): Promise<Article[] | null> {
       const { rows } = await sql`
         SELECT * FROM articles
         WHERE DATE(published_date) = ${date}
         ORDER BY published_date DESC
       `;
       return rows.length > 0 ? rows.map(rowToArticle) : null;
     }

     async saveSummary(summary: StoredSummary): Promise<void> {
       await sql`
         INSERT INTO summaries (id, title, summary, url, source_name, published_date, generated_at)
         VALUES (${summary.id}, ${summary.title}, ${summary.summary}, ${summary.url}, ${summary.sourceName}, ${summary.publishedDate}, ${summary.generatedAt})
         ON CONFLICT (id) DO UPDATE SET
           summary = EXCLUDED.summary,
           generated_at = EXCLUDED.generated_at
       `;
     }

     async getSummary(id: string): Promise<StoredSummary | null> {
       const { rows } = await sql`SELECT * FROM summaries WHERE id = ${id}`;
       return rows.length > 0 ? rowToSummary(rows[0]) : null;
     }

     async saveDigest(digest: DailyDigest): Promise<void> {
       await sql`
         INSERT INTO digests (date, articles, summary, created_at)
         VALUES (${digest.date}, ${JSON.stringify(digest.articles)}, ${digest.summary}, ${digest.created})
         ON CONFLICT (date) DO UPDATE SET
           articles = EXCLUDED.articles,
           summary = EXCLUDED.summary,
           created_at = EXCLUDED.created_at
       `;
     }

     async getDigest(date: string): Promise<DailyDigest | null> {
       const { rows } = await sql`SELECT * FROM digests WHERE date = ${date}`;
       return rows.length > 0 ? rowToDigest(rows[0]) : null;
     }

     async getLatestDigest(): Promise<DailyDigest | null> {
       const { rows } = await sql`
         SELECT * FROM digests
         ORDER BY date DESC
         LIMIT 1
       `;
       return rows.length > 0 ? rowToDigest(rows[0]) : null;
     }

     async getAllSummaries(): Promise<StoredSummary[]> {
       const { rows } = await sql`SELECT * FROM summaries ORDER BY generated_at DESC`;
       return rows.map(rowToSummary);
     }
   }

   // Helper functions to convert DB rows to TypeScript objects
   function rowToArticle(row: any): Article {
     return {
       id: row.id,
       title: row.title,
       summary: row.summary,
       originalSummary: row.original_summary,
       url: row.url,
       sourceName: row.source_name,
       sourceId: row.source_id,
       publishedDate: row.published_date,
       imageUrl: row.image_url,
       categories: row.categories,
     };
   }

   function rowToSummary(row: any): StoredSummary {
     return {
       id: row.id,
       title: row.title,
       summary: row.summary,
       url: row.url,
       sourceName: row.source_name,
       publishedDate: row.published_date,
       generatedAt: row.generated_at,
     };
   }

   function rowToDigest(row: any): DailyDigest {
     return {
       date: row.date,
       articles: row.articles,
       summary: row.summary,
       created: row.created_at,
     };
   }
   ```

5. Update storage instantiation:
   ```typescript
   // Before (file-based):
   import { FileStorage } from '@/lib/storage/file-storage';
   const storage = new FileStorage('./data');

   // After (Postgres):
   import { PostgresStorage } from '@/lib/storage/postgres-storage';
   const storage = new PostgresStorage();
   ```

#### Option B: Vercel KV (Redis)

**Pros:**
- Extremely fast
- Simple key-value storage
- Native Vercel integration

**Cons:**
- Less structured than SQL
- Harder to query/filter

**Setup:**
```bash
npm install @vercel/kv
```

#### Option C: External Database (Supabase, PlanetScale)

**Pros:**
- More control
- Can access outside Vercel
- Generous free tiers

**Cons:**
- Slightly more complex setup
- External service dependency

---

### 2. Set Up Vercel Cron Jobs for Digest Generation

Vercel Cron allows scheduled function execution.

#### Create Cron API Route

**File:** `app/api/cron/generate-digest/route.ts`

```typescript
import { NextRequest, NextResponse } from 'next/server';
import { PostgresStorage } from '@/lib/storage/postgres-storage';
// Import digest generation logic from scripts/generate-digest.ts

export async function GET(request: NextRequest) {
  // Verify cron secret to prevent unauthorized access
  const authHeader = request.headers.get('authorization');
  if (authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const storage = new PostgresStorage();

    // Run digest generation logic
    // (move logic from scripts/generate-digest.ts into a reusable function)
    await generateDailyDigest(storage);

    return NextResponse.json({
      success: true,
      message: 'Digest generated successfully'
    });
  } catch (error) {
    console.error('Cron job error:', error);
    return NextResponse.json({
      error: 'Failed to generate digest'
    }, { status: 500 });
  }
}
```

#### Configure Vercel Cron

**File:** `vercel.json`

```json
{
  "crons": [
    {
      "path": "/api/cron/generate-digest",
      "schedule": "0 6 * * *"
    }
  ]
}
```

**Schedule formats:**
- `"0 6 * * *"` - Daily at 6:00 AM UTC
- `"0 */4 * * *"` - Every 4 hours
- `"0 12 * * 1"` - Every Monday at 12:00 PM UTC

#### Add CRON_SECRET Environment Variable

In Vercel dashboard → Settings → Environment Variables:
```
CRON_SECRET=your-random-secret-string-here
```

---

### 3. Environment Variables

Configure in Vercel dashboard (Settings → Environment Variables):

```bash
# Required
CLAUDE_API_KEY=sk-ant-...
NEWSAPI_KEY=your-newsapi-key
CRON_SECRET=your-cron-secret

# Optional
CLAUDE_MODEL=claude-opus-4-6
POSTGRES_URL=postgresql://...  # Auto-populated by Vercel Postgres
```

---

## Deployment Steps

### 1. Prepare Repository

```bash
# Ensure code is committed
git add .
git commit -m "Prepare for Vercel deployment"
git push origin main
```

### 2. Connect to Vercel

1. Go to [vercel.com](https://vercel.com)
2. Click "New Project"
3. Import your Git repository
4. Vercel will auto-detect Next.js

### 3. Configure Project

- **Framework Preset:** Next.js
- **Root Directory:** `./` (unless you have a monorepo)
- **Build Command:** `npm run build` (auto-detected)
- **Output Directory:** `.next` (auto-detected)

### 4. Add Vercel Postgres

1. In your Vercel project → Storage tab
2. Click "Create Database"
3. Select "Postgres"
4. Create database (free tier available)
5. Run SQL schema from section 1 above

### 5. Set Environment Variables

Add all variables from section 3 above.

### 6. Deploy

Click "Deploy" - Vercel will build and deploy your app.

### 7. Set Up Cron (Post-Deploy)

1. Go to project Settings → Cron Jobs
2. Verify your cron is listed (from `vercel.json`)
3. Test manually: `curl -H "Authorization: Bearer YOUR_CRON_SECRET" https://your-app.vercel.app/api/cron/generate-digest`

### 8. Initial Digest Generation

After first deployment, trigger cron manually to populate database:
```bash
curl -X GET \
  -H "Authorization: Bearer YOUR_CRON_SECRET" \
  https://your-app.vercel.app/api/cron/generate-digest
```

---

## Refactoring Checklist

- [ ] Create PostgresStorage implementation
- [ ] Run database schema setup
- [ ] Update `app/api/digest/route.ts` to use PostgresStorage
- [ ] Update `app/api/summary/[id]/route.ts` to use PostgresStorage
- [ ] Refactor `scripts/generate-digest.ts` into reusable function
- [ ] Create `/api/cron/generate-digest/route.ts`
- [ ] Add `vercel.json` with cron configuration
- [ ] Update `.gitignore` to exclude `.vercel/`
- [ ] Test locally with Vercel CLI: `npx vercel dev`
- [ ] Deploy to Vercel
- [ ] Configure environment variables
- [ ] Set up Vercel Postgres
- [ ] Trigger initial digest generation
- [ ] Verify cron job runs successfully

---

## Local Development with Vercel

Use Vercel CLI to test locally with production-like environment:

```bash
# Install Vercel CLI
npm i -g vercel

# Link to your Vercel project
vercel link

# Pull environment variables
vercel env pull .env.local

# Run dev server with Vercel environment
vercel dev
```

---

## Alternative: Hybrid Approach

Keep file storage for local development, use database for production:

```typescript
// lib/storage/index.ts
import { FileStorage } from './file-storage';
import { PostgresStorage } from './postgres-storage';

export function getStorage() {
  if (process.env.VERCEL) {
    return new PostgresStorage();
  } else {
    return new FileStorage('./data');
  }
}
```

Then use:
```typescript
const storage = getStorage();
```

---

## Cost Estimates (Vercel Free Tier)

- **Hosting:** Free for personal projects
- **Vercel Postgres:** 256 MB free, then $0.30/GB/month
- **Bandwidth:** 100 GB/month free
- **Function Executions:** 100 GB-hours/month free
- **Cron Jobs:** Included in free tier

For this project, free tier should be sufficient for development/personal use.

---

## Troubleshooting

### Functions Timeout
- Increase function timeout in `vercel.json`:
  ```json
  {
    "functions": {
      "app/api/cron/generate-digest/route.ts": {
        "maxDuration": 60
      }
    }
  }
  ```

### Database Connection Errors
- Verify `POSTGRES_URL` is set
- Check database is in same region as functions
- Ensure connection pooling is configured

### Cron Not Running
- Check Vercel dashboard → Deployments → Functions tab
- Verify `CRON_SECRET` matches
- Check function logs for errors

---

## Future Enhancements

- **CDN Caching:** Cache digest responses at edge
- **Image Optimization:** Use Vercel Image Optimization
- **Analytics:** Add Vercel Analytics
- **Preview Deployments:** Test changes before production
- **Edge Functions:** Move read-heavy operations to edge for better performance

---

## Resources

- [Vercel Postgres Documentation](https://vercel.com/docs/storage/vercel-postgres)
- [Vercel Cron Jobs Documentation](https://vercel.com/docs/cron-jobs)
- [Next.js Deployment Documentation](https://nextjs.org/docs/deployment)
- [Vercel Environment Variables](https://vercel.com/docs/concepts/projects/environment-variables)
