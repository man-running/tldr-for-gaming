# TLDR for Gaming - Adaptation Plan

## Overview
Transform the TLDR application from a research paper aggregator to an iGaming news aggregator. The application will fetch articles from iGaming news websites, summarize them using AI, and present a curated "top 5" daily digest.

---

## 1. Current Architecture Analysis

### How TLDR Currently Works
```
Research Paper Sources (ArXiv, HuggingFace)
  ↓
RSS Feed Parser (lib/feed/rssparser.go)
  ↓
Feed Fetcher (lib/feed/getfeed.go) - Fetches & Caches
  ↓
Ranking System - Selects top papers (by relevance/upvotes)
  ↓
AI Summarizer (lib/summary/generator.go) - Uses OpenAI API
  ↓
RSS Output / Web UI Display
```

### Key Components
- **Feed Parsing**: `lib/feed/` - Robust RSS/XML parsing
- **Caching**: Blob storage caching for feed & summaries
- **Summarization**: `lib/summary/` - OpenAI integration
- **API Routes**: `api/` - RESTful endpoints
- **UI**: React/Next.js components
- **Daily Updates**: Scheduled at 7am UTC

---

## 2. Target iGaming News Sources

### Primary Sources (Recommended)
1. **iGamingBusiness.com** - Comprehensive coverage
   - RSS: `https://www.igamingbusiness.com/feed/`
2. **eGamingReview.com** - Industry analysis
   - RSS: May need to confirm endpoint
3. **GamblingInsider.com** - News & events
   - RSS: `https://www.gamblinginsider.com/feed/`
4. **ICE News/Trade** - Conference coverage
   - RSS: Trade publication feeds
5. **Sportech/iGaming News Services**
   - Multiple specialized feeds

### Feed Discovery Strategy
- Use RSS discovery (check for `/feed` or `/rss` endpoints)
- Check for Open Feed Lists (check meta tags, sitemap)
- Contact publishers for official RSS URLs if needed
- Implement RSS auto-discovery from HTML meta tags

---

## 3. Detailed Adaptation Strategy

### Phase 1: Data Model Updates

#### 3.1.1 Create New Types (`lib/article/types.go`)
```go
package article

type ArticleData struct {
    ID            string    // Unique identifier (hash of URL or custom)
    Title         string    // Article headline
    Summary       string    // Original article excerpt/description
    FullContent   string    // Scraped article body (optional)
    URL           string    // Source article URL
    SourceName    string    // e.g., "iGamingBusiness", "Sportech"
    PublishedDate string    // RFC3339 format
    ImageURL      string    // Featured image
    Categories    []string  // Tags: "Regulations", "Sports Betting", "iCasino", etc.
    Authors       []string  // Article author(s)
    Metadata      map[string]interface{} // Extra fields (views, engagement, etc.)
}

type ArticleMetadata struct {
    Title         string
    SourceName    string
    PublishedDate string
    URL           string
    Categories    []string
}

type RankedArticle struct {
    Article ArticleData
    Score   float64      // Ranking score (0-1)
    Reason  string       // Why selected: "trending", "high-engagement", "sector-news"
}
```

#### 3.1.2 Update Constants (`lib/constants.ts`)
```typescript
export const siteConfig = {
  name: "iGaming TLDR",
  description: "Daily curated iGaming & sports betting news summaries",
  url: "https://gaming-tldr.example.com",
  links: {
    newsSources: [
      "https://www.igamingbusiness.com/feed/",
      "https://www.gamblinginsider.com/feed/",
      // ... more sources
    ]
  }
};
```

### Phase 2: Feed Source Management

#### 3.2.1 Create News Feed Configuration (`lib/feed/sources.go`)
```go
package feed

type NewsSource struct {
    ID           string
    Name         string
    URL          string
    FeedURL      string
    Category     string     // "Regulations", "Business", "Technology", etc.
    Active       bool
    Priority     int        // Higher = more important in ranking
    ScrapingType string     // "rss", "scrape", "api"
}

type SourceManager struct {
    sources []NewsSource
}

func (sm *SourceManager) LoadSources() error {
    // Load from config file or environment
}

func (sm *SourceManager) FetchAllFeeds(ctx context.Context) ([]Article, error) {
    // Fetch from all active sources
}
```

#### 3.2.2 Enhance RSS Parser (`lib/feed/rssparser.go`)
- Current parser already handles RSS well
- **Enhancements needed**:
  - Better content extraction (from `<description>`, `<content:encoded>`)
  - Image extraction from media RSS extensions
  - Category/tag extraction
  - Author field handling

### Phase 3: Content Enrichment

#### 3.3.1 Create Web Scraper (`lib/feed/scraper.go`)
For sources without RSS feeds, implement:
```go
package feed

type Scraper interface {
    Scrape(url string) (*Article, error)
}

// Implement for specific sites:
// - Parse HTML using goquery/colly
// - Extract: title, body, publish date, image
// - Handle pagination
```

#### 3.3.2 Update Summarization (`lib/summary/generator.go`)
- **Current implementation**: Works with research papers
- **Adapt for news articles**:
  - Remove paper-specific prompts
  - Use optimized prompt for news summarization
  - Keep similar OpenAI integration
  - Preserve caching mechanism

**Example prompt change**:
```
BEFORE:
"Summarize this academic research paper in 2-3 sentences..."

AFTER:
"Summarize this iGaming industry news article in 2-3 sentences,
highlighting key business implications and regulatory impacts..."
```

### Phase 4: Ranking & Selection Logic

#### 3.4.1 Create Ranking Engine (`lib/article/ranker.go`)
```go
package article

type RankingCriteria struct {
    RecencyWeight    float64  // Recent articles score higher
    SourceWeight     float64  // Trusted sources score higher
    CategoryWeight   float64  // Diversify across categories
    EngagementWeight float64  // Comments, shares (if available)
}

type Ranker struct {
    criteria RankingCriteria
}

func (r *Ranker) RankArticles(articles []ArticleData) []RankedArticle {
    // Score each article
    // Apply weighting
    // Sort and select top 5
    // Ensure category diversity
}

func (r *Ranker) EnsureDiversity(articles []RankedArticle) []RankedArticle {
    // Ensure we have articles from different:
    // - Sources
    // - Categories/sectors
    // - Geographic regions (if relevant)
}
```

**Ranking Factors**:
- **Recency** (40%): Published in last 24 hours
- **Source Authority** (30%): Weight by source importance
- **Engagement** (20%): Comments, shares (if available)
- **Category Diversity** (10%): Spread across sectors

### Phase 5: API & Endpoint Updates

#### 3.5.1 Modify API Routes

**Keep existing patterns**, update endpoints:

| Current | New | Purpose |
|---------|-----|---------|
| `/api/feed` | `/api/news-feed` | Get raw articles feed |
| `/api/papers` | `/api/articles` | Get processed articles |
| `/api/tldr` | `/api/news-tldr` | Get daily summary with AI summaries |
| `/api/paper/[id]` | `/api/article/[id]` | Get single article details |
| `/api/search` | `/api/search` | Search articles |

#### 3.5.2 Update Response Types
```typescript
// From paper-centric:
{
  title: string,
  abstract: string,
  arxivId: string,
  authors: string[]
}

// To news-centric:
{
  id: string,
  title: string,
  summary: string,           // AI-generated summary
  originalSummary: string,   // Article excerpt
  source: string,
  url: string,
  publishedDate: string,
  category: string,
  image: string,
  ranking: { score: number, reason: string }
}
```

### Phase 6: UI Component Updates

#### 3.6.1 Update React Components
The component structure can remain mostly the same, with content updates:

**Files to update**:
- `components/feed/` - Change paper rendering to article rendering
- `components/paper/` → `components/article/` - Rename folder
  - `abstract-renderer.tsx` - Show article body instead
  - Update link handling (URL vs internal paper view)
- `app/p/[arxivId]/` → `app/article/[id]/` - Single article view
- `components/layout/` - Update header/footer text

**Key changes**:
- Replace arxivId with URL-based routing
- Add source/category badges
- Update button labels ("View Paper" → "Read Full Article")
- Add sharing buttons for news articles
- Update "favorites" to "saved articles"

#### 3.6.2 Update UI Styling
- Change colors/theme for gaming industry (consider: professional gaming colors)
- Update logo/branding
- Adjust typography if needed

### Phase 7: Database & Storage

#### 3.7.1 Update Schema
```sql
-- Instead of:
CREATE TABLE papers (
  arxiv_id VARCHAR PRIMARY KEY,
  title TEXT,
  ...
);

-- Use:
CREATE TABLE articles (
  id VARCHAR PRIMARY KEY,
  source_id VARCHAR,
  url TEXT UNIQUE,
  title TEXT,
  original_summary TEXT,
  ai_summary TEXT,
  category VARCHAR,
  published_date TIMESTAMP,
  created_at TIMESTAMP,
  ...
);

-- Track sources:
CREATE TABLE news_sources (
  id VARCHAR PRIMARY KEY,
  name VARCHAR,
  feed_url TEXT,
  category VARCHAR,
  active BOOLEAN,
  ...
);

-- Track rankings:
CREATE TABLE article_rankings (
  article_id VARCHAR,
  date DATE,
  rank INT,
  score FLOAT,
  reason VARCHAR,
  PRIMARY KEY (article_id, date)
);
```

### Phase 8: Configuration & Environment

#### 3.8.1 Environment Variables
```bash
# News Sources
IGAMING_NEWS_SOURCES="igamingbusiness,gamblinginsider,sportech"

# Summarization
OPENAI_API_KEY=xxx
SUMMARY_PROMPT_TEMPLATE="iGaming"

# Ranking
RANKING_UPDATE_TIME="07:00"  # UTC
RANKING_CRITERIA="recency,source,engagement,diversity"

# Database (if using for historical data)
DATABASE_URL=xxx
ENABLE_ARTICLE_HISTORY=true

# Caching
CACHE_TTL=300  # 5 minutes

# Scraping
ENABLE_WEB_SCRAPING=true
SCRAPING_TIMEOUT=10000  # ms
```

---

## 4. Implementation Roadmap

### Phase 1: Foundation (Week 1)
- [ ] Create new Go types for articles
- [ ] Set up news feed sources configuration
- [ ] Create source manager
- [ ] Update constants and branding

### Phase 2: Feed Integration (Week 2)
- [ ] Enhance RSS parser for news-specific fields
- [ ] Implement basic article fetcher
- [ ] Test with 2-3 iGaming sources
- [ ] Set up caching for articles

### Phase 3: Summarization (Week 3)
- [ ] Adapt summarization prompts
- [ ] Test with sample articles
- [ ] Optimize for news content
- [ ] Implement fallback (use article excerpt if summary fails)

### Phase 4: Ranking & Selection (Week 3)
- [ ] Implement ranking algorithm
- [ ] Add diversity constraints
- [ ] Create daily top-5 selection
- [ ] Test ranking logic

### Phase 5: API Updates (Week 4)
- [ ] Update/rename API endpoints
- [ ] Create new response types
- [ ] Update OpenAPI docs
- [ ] Add migration layer if needed

### Phase 6: Frontend Updates (Week 4-5)
- [ ] Update React components
- [ ] Rename internal folders
- [ ] Update UI text/branding
- [ ] Test all pages

### Phase 7: Testing & Optimization (Week 5)
- [ ] End-to-end testing
- [ ] Performance testing
- [ ] Feed reliability testing
- [ ] Daily scheduler testing

### Phase 8: Launch Preparation (Week 6)
- [ ] Deploy to staging
- [ ] Production configuration
- [ ] Monitoring setup
- [ ] Documentation updates

---

## 5. Data Flow Diagram

```
iGaming News Sources (Multiple RSS Feeds)
         ↓
Source Manager (Parallel Fetching)
         ↓
Feed Parser (Extract articles)
         ↓
Article Enrichment (Add metadata)
         ↓
Caching Layer (5-min TTL)
         ↓
Ranking Engine (Score & select top 5)
         ↓
Summarization Service (OpenAI)
         ↓
Cache Summary Results
         ↓
API Endpoints
    ├─ /api/news-feed (raw articles)
    ├─ /api/articles (processed)
    ├─ /api/news-tldr (daily top 5)
    └─ /api/article/[id] (single article)
         ↓
React Frontend
    ├─ Feed View (all articles)
    ├─ Daily Summary (top 5)
    ├─ Article Detail Page
    └─ Search/Filter
```

---

## 6. Key Considerations

### 6.1 RSS Feed Reliability
- Some iGaming sites may not have public RSS feeds
- **Solution**:
  - Maintain list of available feeds
  - Implement web scraper for major sites without RSS
  - Add fallback mechanisms

### 6.2 Content Freshness
- News moves quickly in iGaming (regulatory changes, etc.)
- **Solution**:
  - Increase fetch frequency (every 2-4 hours instead of daily)
  - Maintain real-time feed, but curate daily top-5

### 6.3 Summarization Quality
- News articles vary greatly in structure
- **Solution**:
  - Fine-tune prompts for different article types
  - Implement quality checks
  - Allow fallback to article excerpt

### 6.4 Source Quality & Bias
- Ensure balance across sources
- **Solution**:
  - Weight by established reputation
  - Monitor for duplicate coverage
  - Ensure geographical diversity (UK, US, EU, Asia)

### 6.5 Regulatory Compliance
- iGaming is heavily regulated
- **Solution**:
  - Add category filtering by jurisdiction
  - Track regulatory impact level
  - Add disclaimers/context where needed

---

## 7. Testing Strategy

### Unit Tests
- Ranking algorithm
- Feed parsing
- Article enrichment
- Summarization prompt handling

### Integration Tests
- Multi-source feed fetching
- End-to-end article pipeline
- API endpoints
- Caching mechanisms

### E2E Tests
- User flows (viewing feed, reading article, etc.)
- Daily update process
- Email subscription (if implemented)

---

## 8. Future Enhancements

1. **Real-time Notifications**: Alert users about breaking news
2. **Personalization**: User preferences for sources/categories
3. **Email Digest**: Daily email with top 5 summaries
4. **Mobile App**: Native iOS/Android app
5. **Search Enhancement**: Full-text search with filters
6. **Analytics**: Track what's trending in iGaming
7. **API for Third Parties**: Allow integration with other platforms
8. **Multi-language**: Translate summaries for different markets

---

## 9. File Structure Changes Summary

```
Current (TLDR):
lib/
├── paper/
├── summary/
└── feed/

New (iGaming):
lib/
├── article/            # NEW: Article-specific logic
├── article/types.go    # NEW: Article data structures
├── article/ranker.go   # NEW: Ranking algorithm
├── feed/
│   ├── sources.go      # NEW: Feed source management
│   ├── scraper.go      # NEW: Web scraper for non-RSS sites
│   └── (enhanced existing files)
└── summary/
    └── (adapted existing generator)

api/
├── news-feed/          # NEW: Raw feed endpoint
├── articles/           # NEW/RENAMED: From /papers
├── news-tldr/          # RENAMED: From /tldr
├── article/            # NEW/RENAMED: Single article endpoint
└── search/             # NEW/RENAMED: Article search

components/
├── article/            # NEW/RENAMED: From /paper
├── feed/               # (existing, minimal changes)
└── layout/             # (minimal changes)

app/
├── article/            # NEW: Article detail pages
└── (rename /p to /article if needed)
```

---

## 10. Success Metrics

- ✅ Successfully fetch from 3+ iGaming news sources
- ✅ Parse and summarize 50+ articles daily
- ✅ Select balanced top 5 daily
- ✅ AI summaries match article quality
- ✅ Page load time < 2 seconds
- ✅ API response time < 500ms
- ✅ 99.9% feed availability
- ✅ UI/UX matches current TLDR quality

---

## 11. Next Steps

1. Review and approve this plan
2. Prioritize which phases to tackle first
3. Set up development branch: `feature/igaming-adaptation`
4. Start Phase 1: Create new data types
5. Test with one live iGaming news source

Would you like me to proceed with implementing any specific phase?
