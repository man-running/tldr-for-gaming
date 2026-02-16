/**
 * Digest Generation Module
 * Reusable functions for generating daily iGaming news digests
 */

import { fetchArticles } from './article-fetcher';
import type { IStorage, Article, StoredSummary, RankedArticle, DailyDigest } from '../storage/storage-interface';

/**
 * Generate Claude summary for individual article
 */
export async function generateArticleSummary(article: Article): Promise<string> {
  const apiKey = process.env.CLAUDE_API_KEY;

  if (!apiKey) {
    console.warn('CLAUDE_API_KEY not set, using original summary');
    return article.summary || article.originalSummary;
  }

  try {
    const response = await fetch('https://api.anthropic.com/v1/messages', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': apiKey,
        'anthropic-version': '2023-06-01',
      },
      body: JSON.stringify({
        model: process.env.CLAUDE_MODEL || 'claude-opus-4-6',
        max_tokens: 300,
        messages: [
          {
            role: 'user',
            content: `Write a pithy, concise summary of this iGaming/online gambling industry article. Get straight to the point.

Title: ${article.title}
Source: ${article.sourceName}
Original Summary: ${article.summary || article.originalSummary}

Write 1-2 short paragraphs that capture the key facts and why this matters for the online gambling industry. Be concise - focus on what's new, what's changing, and what it means. Skip unnecessary context.

IMPORTANT: Do NOT include a title or heading (no ## or # markdown). Start directly with the content paragraph.`,
          },
        ],
      }),
    });

    if (!response.ok) {
      console.error('Claude API error for article summary:', await response.text());
      return article.summary || article.originalSummary;
    }

    const data = await response.json();
    const textContent = data.content.find(
      (c: { type: string }) => c.type === 'text'
    );
    return textContent?.text || article.summary || article.originalSummary;
  } catch (error) {
    console.error('Error generating article summary:', error);
    return article.summary || article.originalSummary;
  }
}

/**
 * Generate narrative summary with Claude API
 */
export async function generateNarrativeSummary(articles: Article[]): Promise<string> {
  const apiKey = process.env.CLAUDE_API_KEY;

  if (!apiKey) {
    return generateTemplateSummary(articles);
  }

  try {
    const response = await fetch('https://api.anthropic.com/v1/messages', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': apiKey,
        'anthropic-version': '2023-06-01',
      },
      body: JSON.stringify({
        model: process.env.CLAUDE_MODEL || 'claude-opus-4-6',
        max_tokens: 800,
        messages: [
          {
            role: 'user',
            content: `You are a professional news digest writer specializing in the global iGaming and online gambling industry. Write a concise 2-3 paragraph narrative summary of today's TOP 5 iGaming industry news stories from a GLOBAL perspective (not US-centric). Weave the following article titles naturally into the text as hyperlinks (use markdown format [Title](url)). Make it flow like a narrative, not a list.

Top 5 Articles:
${articles
  .map((a) => `[${a.title}](${a.url}) - ${a.summary || a.originalSummary}`)
  .join('\n')}

Write a compelling, concise narrative that connects these 5 stories and highlights key themes. Focus on international markets, regulations, and operators. Keep it brief and focused - 2-3 paragraphs maximum. Do not default to a US perspective - treat all regions equally.`,
          },
        ],
      }),
    });

    if (!response.ok) {
      console.error('Claude API error:', await response.text());
      return generateTemplateSummary(articles);
    }

    const data = await response.json();
    const textContent = data.content.find(
      (c: { type: string }) => c.type === 'text'
    );
    return textContent?.text || generateTemplateSummary(articles);
  } catch (error) {
    console.error('Error calling Claude API:', error);
    return generateTemplateSummary(articles);
  }
}

/**
 * Generate headline using Claude API
 */
export async function generateHeadline(articles: Article[]): Promise<string> {
  const apiKey = process.env.CLAUDE_API_KEY;

  if (!apiKey) {
    return generateTemplateHeadline(articles);
  }

  try {
    const response = await fetch('https://api.anthropic.com/v1/messages', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': apiKey,
        'anthropic-version': '2023-06-01',
      },
      body: JSON.stringify({
        model: process.env.CLAUDE_MODEL || 'claude-opus-4-6',
        max_tokens: 50,
        messages: [
          {
            role: 'user',
            content: `Write a single, compelling headline (one sentence, max 15 words) that captures the main theme of today's iGaming news. Be specific and newsworthy.

Top stories:
${articles.map((a, i) => `${i + 1}. ${a.title}`).join('\n')}

Write just the headline - no quotes, no title, no formatting. Keep it under 15 words.`,
          },
        ],
      }),
    });

    if (!response.ok) {
      console.error('Claude API error for headline:', await response.text());
      return generateTemplateHeadline(articles);
    }

    const data = await response.json();
    const textContent = data.content.find(
      (c: { type: string }) => c.type === 'text'
    );
    return textContent?.text || generateTemplateHeadline(articles);
  } catch (error) {
    console.error('Error generating headline:', error);
    return generateTemplateHeadline(articles);
  }
}

/**
 * Template-based fallback headline
 */
export function generateTemplateHeadline(articles: Article[]): string {
  if (articles.length === 0) {
    return 'No major stories today';
  }
  return `Top story: ${articles[0].title}`;
}

/**
 * Template-based fallback summary
 */
export function generateTemplateSummary(articles: Article[]): string {
  if (articles.length < 3) {
    return 'Not enough articles for summary';
  }

  const [a1, a2, a3, a4, a5] = articles;

  return `Big news in iGaming: [${a1.title}](${a1.url}) ${a1.summary || a1.originalSummary}

Meanwhile, [${a2.title}](${a2.url}) ${a2.summary || a2.originalSummary} Industry experts see this as a major shift in the online gambling landscape.

On the regulatory front, [${a3.title}](${a3.url}) ${a3.summary || a3.originalSummary}

The technology front is also heating up as [${a4.title}](${a4.url}) ${a4.summary || a4.originalSummary} This reflects a broader industry trend toward innovation.

Finally, [${a5.title}](${a5.url}) ${a5.summary || a5.originalSummary} The iGaming industry continues to expand in new directions.`;
}

/**
 * Main digest generation function - accepts storage interface
 */
export async function generateDailyDigest(storage: IStorage): Promise<DailyDigest> {
  console.log('\n=== Starting Digest Generation ===\n');

  // Initialize storage
  await storage.initialize();

  // Get today's date
  const today = new Date().toISOString().split('T')[0];
  console.log(`Generating digest for: ${today}`);

  // Step 1: Fetch articles
  console.log('\n[1/5] Fetching articles from NewsAPI...');
  const articles = await fetchArticles();

  if (articles.length === 0) {
    throw new Error('No articles fetched');
  }

  console.log(`✓ Fetched ${articles.length} articles`);

  // Step 2: Save raw articles
  console.log('\n[2/5] Saving raw articles...');
  await storage.saveArticles(today, articles);
  console.log('✓ Articles saved');

  // Step 3: Generate and save summaries for ALL articles (so narrative links work)
  console.log('\n[3/5] Generating Claude summaries for all articles...');
  const articlesToSummarize = articles.slice(0, 15); // Generate summaries for top 15

  for (let i = 0; i < articlesToSummarize.length; i++) {
    const article = articlesToSummarize[i];
    console.log(`  Generating summary ${i + 1}/${articlesToSummarize.length}: ${article.title.substring(0, 60)}...`);

    const summary = await generateArticleSummary(article);

    const storedSummary: StoredSummary = {
      id: article.id,
      title: article.title,
      summary,
      url: article.url,
      sourceName: article.sourceName,
      publishedDate: article.publishedDate,
      generatedAt: new Date().toISOString(),
      imageUrl: article.imageUrl,
    };

    await storage.saveSummary(storedSummary);
  }

  console.log(`✓ All ${articlesToSummarize.length} summaries generated and saved`);

  // Step 4: Create digest with summary page links
  console.log('\n[4/5] Creating digest...');

  // Update article URLs to point to summary pages
  const articlesWithSummaryLinks = articles.map((article) => ({
    ...article,
    url: `/summary/${article.id}`,
  }));

  // Create ranked articles (top 5)
  const topFiveArticles = articlesWithSummaryLinks.slice(0, 5);
  const rankedArticles: RankedArticle[] = topFiveArticles
    .map((article, index) => ({
      article,
      rank: index + 1,
      score: 0.95 - index * 0.08,
      reason: 'relevant',
    }));

  // Generate headline and narrative summary (only for top 5)
  const headline = await generateHeadline(topFiveArticles);
  let narrativeSummary = await generateNarrativeSummary(topFiveArticles);

  // Replace direct article URLs with summary page URLs (top 5 only)
  for (let i = 0; i < topFiveArticles.length; i++) {
    const article = articles[i];
    const summaryArticle = topFiveArticles[i];
    narrativeSummary = narrativeSummary.split(article.url).join(summaryArticle.url);
  }

  const digest: DailyDigest = {
    date: today,
    articles: rankedArticles,
    headline: headline,
    summary: narrativeSummary,
    created: new Date().toISOString(),
  };

  console.log('✓ Digest created');

  // Step 5: Save digest
  console.log('\n[5/5] Saving digest...');
  await storage.saveDigest(digest);
  console.log('✓ Digest saved');

  console.log('\n=== Digest Generation Complete! ===\n');
  console.log(`Date: ${today}`);
  console.log(`Articles: ${articles.length}`);
  console.log(`Summaries: ${articlesToSummarize.length}`);

  return digest;
}
