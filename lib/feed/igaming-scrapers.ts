/**
 * iGaming News Scrapers
 * Fetches articles from specialized iGaming industry news sources via RSS feeds
 */

import Parser from 'rss-parser';
import type { Article } from './article-fetcher';

const parser = new Parser({
  timeout: 10000,
  headers: {
    'User-Agent': 'Mozilla/5.0 (compatible; TLDRGaming/1.0)',
  },
});

/**
 * Generate a stable ID from a URL
 */
function generateStableId(url: string, source: string): string {
  let hash = 0;
  for (let i = 0; i < url.length; i++) {
    const char = url.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash;
  }
  return `${source}-${Math.abs(hash)}`;
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
    "Mobile Gaming": ["mobile casino", "mobile betting", "app"],
    Regulation: ["regulation", "license", "regulated", "compliance", "legal"],
    Technology: ["technology", "platform", "software", "innovation"],
    "M&A": ["merger", "acquisition", "acquired", "investment", "funding"],
    Industry: ["operator", "market", "business", "company"],
  };

  for (const [category, keywords] of Object.entries(categoryKeywords)) {
    if (keywords.some((keyword) => lowerText.includes(keyword))) {
      categories.push(category);
    }
  }

  return categories.length > 0 ? categories : ["iGaming"];
}

/**
 * Fetch articles from SBC News (Sports Betting Community)
 * RSS: https://www.sbcnews.co.uk/feed/
 */
export async function fetchFromSBCNews(): Promise<Article[]> {
  try {
    console.log('Fetching from SBC News...');
    const feed = await parser.parseURL('https://www.sbcnews.co.uk/feed/');

    const articles: Article[] = feed.items.slice(0, 10).map((item) => ({
      id: generateStableId(item.link || '', 'sbc'),
      title: item.title || 'Untitled',
      summary: item.contentSnippet?.substring(0, 300) || item.content?.substring(0, 300) || '',
      originalSummary: item.contentSnippet?.substring(0, 300) || '',
      url: item.link || '',
      sourceName: 'SBC News',
      sourceId: 'sbc-news',
      publishedDate: new Date(item.pubDate || Date.now()).toISOString(),
      categories: inferCategories((item.title || '') + ' ' + (item.contentSnippet || '')),
    }));

    console.log(`✓ Fetched ${articles.length} articles from SBC News`);
    return articles;
  } catch (error) {
    console.error('Error fetching from SBC News:', error);
    return [];
  }
}

/**
 * Fetch articles from iGaming Business
 * RSS: https://igamingbusiness.com/feed/
 */
export async function fetchFromIGamingBusiness(): Promise<Article[]> {
  try {
    console.log('Fetching from iGaming Business...');
    const feed = await parser.parseURL('https://igamingbusiness.com/feed/');

    const articles: Article[] = feed.items.slice(0, 10).map((item) => ({
      id: generateStableId(item.link || '', 'igb'),
      title: item.title || 'Untitled',
      summary: item.contentSnippet?.substring(0, 300) || item.content?.substring(0, 300) || '',
      originalSummary: item.contentSnippet?.substring(0, 300) || '',
      url: item.link || '',
      sourceName: 'iGaming Business',
      sourceId: 'igaming-business',
      publishedDate: new Date(item.pubDate || Date.now()).toISOString(),
      categories: inferCategories((item.title || '') + ' ' + (item.contentSnippet || '')),
    }));

    console.log(`✓ Fetched ${articles.length} articles from iGaming Business`);
    return articles;
  } catch (error) {
    console.error('Error fetching from iGaming Business:', error);
    return [];
  }
}

/**
 * Fetch articles from CalvinAyre
 * RSS: https://calvinayre.com/feed/
 */
export async function fetchFromCalvinAyre(): Promise<Article[]> {
  try {
    console.log('Fetching from CalvinAyre...');
    const feed = await parser.parseURL('https://calvinayre.com/feed/');

    const articles: Article[] = feed.items.slice(0, 10).map((item) => ({
      id: generateStableId(item.link || '', 'ca'),
      title: item.title || 'Untitled',
      summary: item.contentSnippet?.substring(0, 300) || item.content?.substring(0, 300) || '',
      originalSummary: item.contentSnippet?.substring(0, 300) || '',
      url: item.link || '',
      sourceName: 'CalvinAyre',
      sourceId: 'calvinayre',
      publishedDate: new Date(item.pubDate || Date.now()).toISOString(),
      categories: inferCategories((item.title || '') + ' ' + (item.contentSnippet || '')),
    }));

    console.log(`✓ Fetched ${articles.length} articles from CalvinAyre`);
    return articles;
  } catch (error) {
    console.error('Error fetching from CalvinAyre:', error);
    return [];
  }
}

/**
 * Fetch articles from EGR Global
 * RSS: https://egr.global/feed/
 */
export async function fetchFromEGR(): Promise<Article[]> {
  try {
    console.log('Fetching from EGR Global...');
    const feed = await parser.parseURL('https://egr.global/feed/');

    const articles: Article[] = feed.items.slice(0, 10).map((item) => ({
      id: generateStableId(item.link || '', 'egr'),
      title: item.title || 'Untitled',
      summary: item.contentSnippet?.substring(0, 300) || item.content?.substring(0, 300) || '',
      originalSummary: item.contentSnippet?.substring(0, 300) || '',
      url: item.link || '',
      sourceName: 'EGR Global',
      sourceId: 'egr-global',
      publishedDate: new Date(item.pubDate || Date.now()).toISOString(),
      categories: inferCategories((item.title || '') + ' ' + (item.contentSnippet || '')),
    }));

    console.log(`✓ Fetched ${articles.length} articles from EGR Global`);
    return articles;
  } catch (error) {
    console.error('Error fetching from EGR Global:', error);
    return [];
  }
}

/**
 * Fetch all articles from iGaming sources
 */
export async function fetchFromIGamingSources(): Promise<Article[]> {
  console.log('\n--- Fetching from iGaming Sources ---');

  // Fetch from all sources in parallel
  const [sbcArticles, igbArticles, caArticles, egrArticles] = await Promise.all([
    fetchFromSBCNews(),
    fetchFromIGamingBusiness(),
    fetchFromCalvinAyre(),
    fetchFromEGR(),
  ]);

  // Combine all articles
  const allArticles = [
    ...sbcArticles,
    ...igbArticles,
    ...caArticles,
    ...egrArticles,
  ];

  // Remove duplicates by URL
  const uniqueArticles = Array.from(
    new Map(allArticles.map((item) => [item.url, item])).values()
  );

  // Sort by published date (most recent first)
  const sortedArticles = uniqueArticles.sort(
    (a, b) =>
      new Date(b.publishedDate).getTime() - new Date(a.publishedDate).getTime()
  );

  console.log(`\n✓ Total unique articles from iGaming sources: ${sortedArticles.length}`);
  return sortedArticles;
}
