/**
 * Article Fetcher - Handles fetching articles from various sources
 * Supports: NewsAPI, direct site scrapers (to be added)
 */

/**
 * Generate a stable ID from a URL
 */
function generateStableId(url: string): string {
  // Simple hash function for URL
  let hash = 0;
  for (let i = 0; i < url.length; i++) {
    const char = url.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32-bit integer
  }
  return `newsapi-${Math.abs(hash)}`;
}

export interface Article {
  id: string;
  title: string;
  summary: string;
  originalSummary: string;
  url: string;
  sourceName: string;
  sourceId: string;
  publishedDate: string;
  imageUrl?: string;
  categories?: string[];
}

/**
 * Fetch gaming news articles from NewsAPI with timeout
 */
export async function fetchFromNewsAPI(): Promise<Article[]> {
  const apiKey = process.env.NEWSAPI_KEY;

  if (!apiKey) {
    console.warn("NEWSAPI_KEY not configured");
    return [];
  }

  try {
    console.log("Fetching articles from NewsAPI...");

    const url = new URL("https://newsapi.org/v2/everything");
    url.searchParams.append("q", "(igaming OR \"online gambling\" OR \"online casino\" OR \"casino operator\") AND -promo -\"bet now\" -\"sign up\"");
    url.searchParams.append("sortBy", "publishedAt");
    url.searchParams.append("language", "en");
    url.searchParams.append("pageSize", "10");
    url.searchParams.append("apiKey", apiKey);

    // Fetch with 10 second timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000);

    const response = await fetch(url.toString(), {
      signal: controller.signal,
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      console.error(`NewsAPI error: ${response.status} ${response.statusText}`);
      const errorData = await response.json().catch(() => null);
      console.error("Error details:", errorData);
      return [];
    }

    const data = await response.json();
    console.log(`NewsAPI returned ${data.articles?.length || 0} articles`);

    if (!data.articles || !Array.isArray(data.articles)) {
      console.warn("No articles in NewsAPI response");
      return [];
    }

    const articles = data.articles
      .filter((article: any) => article.title && article.url)
      .map((article: any) => ({
        id: generateStableId(article.url),
        title: article.title,
        summary: article.description || article.content?.substring(0, 300) || "",
        originalSummary:
          article.description || article.content?.substring(0, 300) || "",
        url: article.url,
        sourceName: article.source?.name || "Gaming News",
        sourceId: article.source?.id || "gaming-news",
        publishedDate: new Date(article.publishedAt).toISOString(),
        imageUrl: article.urlToImage,
        categories: inferCategories(
          article.title + " " + (article.description || "")
        ),
      }));

    // Remove duplicates by URL
    const uniqueArticles = Array.from(
      new Map(articles.map((item: Article) => [item.url, item])).values()
    ) as Article[];

    const result = uniqueArticles
      .sort(
        (a, b) =>
          new Date(b.publishedDate).getTime() -
          new Date(a.publishedDate).getTime()
      )
      .slice(0, 10);

    console.log(`Returning ${result.length} unique articles`);
    return result;
  } catch (error) {
    if (error instanceof Error) {
      if (error.name === "AbortError") {
        console.error("NewsAPI fetch timeout");
      } else {
        console.error("Error fetching from NewsAPI:", error.message);
      }
    } else {
      console.error("Error fetching from NewsAPI:", error);
    }
    return [];
  }
}

/**
 * Infer categories from article title and description
 */
function inferCategories(text: string): string[] {
  const lowerText = text.toLowerCase();
  const categories: string[] = [];

  const categoryKeywords: Record<string, string[]> = {
    "Online Casino": ["casino", "slots", "roulette", "blackjack", "poker"],
    "Sports Betting": ["sports betting", "sportsbook", "betting", "odds", "wager"],
    "iGaming": ["igaming", "online gambling", "gambling platform"],
    "Mobile Gaming": ["mobile casino", "mobile betting", "app", "ios", "android"],
    Regulation: ["regulation", "license", "regulated", "compliance", "legal"],
    Technology: ["technology", "platform", "software", "innovation"],
    Industry: ["industry", "market", "business", "company", "operator"],
    News: ["announcement", "launches", "release", "partnership"],
  };

  for (const [category, keywords] of Object.entries(categoryKeywords)) {
    if (keywords.some((keyword) => lowerText.includes(keyword))) {
      categories.push(category);
    }
  }

  return categories.length > 0 ? categories : ["iGaming"];
}

/**
 * Main function to fetch articles from configured sources
 */
export async function fetchArticles(): Promise<Article[]> {
  console.log("Starting article fetch...");

  // Import iGaming scrapers dynamically to avoid circular dependencies
  const { fetchFromIGamingSources } = await import('./igaming-scrapers');

  // Fetch from both NewsAPI and iGaming sources in parallel
  const [newsApiArticles, iGamingArticles] = await Promise.all([
    fetchFromNewsAPI(),
    fetchFromIGamingSources(),
  ]);

  console.log(`NewsAPI: ${newsApiArticles.length} articles`);
  console.log(`iGaming Sources: ${iGamingArticles.length} articles`);

  // Combine all articles
  const allArticles = [...newsApiArticles, ...iGamingArticles];

  // Remove duplicates by URL
  const uniqueArticles = Array.from(
    new Map(allArticles.map((item) => [item.url, item])).values()
  );

  // Sort by published date (most recent first)
  const sortedArticles = uniqueArticles.sort(
    (a, b) =>
      new Date(b.publishedDate).getTime() - new Date(a.publishedDate).getTime()
  );

  console.log(`Fetched ${sortedArticles.length} total unique articles`);
  return sortedArticles.slice(0, 20); // Return top 20 most recent
}
