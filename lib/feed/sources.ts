// TypeScript type definitions and utilities for news sources
// Used for frontend and API integration

export type ScrapingType = "rss" | "scrape" | "api";

export interface NewsSource {
	id: string;
	name: string;
	url: string;
	feedUrl: string;
	category: string;
	active: boolean;
	priority: number; // 1-10
	scrapingType: ScrapingType;
	timeout: number; // milliseconds
}

export interface NewsSourceMetadata {
	id: string;
	name: string;
	category: string;
	active: boolean;
	priority: number;
}

// Default news sources configuration
export const DEFAULT_NEWS_SOURCES: NewsSource[] = [
	{
		id: "igamingbusiness",
		name: "iGamingBusiness",
		url: "https://www.igamingbusiness.com",
		feedUrl: "https://www.igamingbusiness.com/feed/",
		category: "Business",
		active: true,
		priority: 10,
		scrapingType: "rss",
		timeout: 10000,
	},
	{
		id: "gamblinginsider",
		name: "Gambling Insider",
		url: "https://www.gamblinginsider.com",
		feedUrl: "https://www.gamblinginsider.com/feed/",
		category: "Business",
		active: true,
		priority: 9,
		scrapingType: "rss",
		timeout: 10000,
	},
	{
		id: "egamingreview",
		name: "eGaming Review",
		url: "https://www.egamingreview.com",
		feedUrl: "https://www.egamingreview.com/feed/",
		category: "Regulations",
		active: true,
		priority: 8,
		scrapingType: "rss",
		timeout: 10000,
	},
	{
		id: "sportech",
		name: "Sportech",
		url: "https://www.sportech.com",
		feedUrl: "https://www.sportech.com/feed/",
		category: "Sports Betting",
		active: true,
		priority: 7,
		scrapingType: "rss",
		timeout: 10000,
	},
	{
		id: "bettingindustry",
		name: "Betting Industry",
		url: "https://www.bettingindustry.com",
		feedUrl: "https://www.bettingindustry.com/feed/",
		category: "Sports Betting",
		active: true,
		priority: 7,
		scrapingType: "rss",
		timeout: 10000,
	},
];

/**
 * Get a source by ID
 */
export function getSourceById(id: string): NewsSource | undefined {
	return DEFAULT_NEWS_SOURCES.find((source) => source.id === id);
}

/**
 * Get all active sources
 */
export function getActiveSources(): NewsSource[] {
	return DEFAULT_NEWS_SOURCES.filter((source) => source.active).sort(
		(a, b) => b.priority - a.priority
	);
}

/**
 * Get sources by category
 */
export function getSourcesByCategory(category: string): NewsSource[] {
	return DEFAULT_NEWS_SOURCES.filter(
		(source) => source.active && source.category === category
	).sort((a, b) => b.priority - a.priority);
}

/**
 * Get source metadata
 */
export function getSourceMetadata(id: string): NewsSourceMetadata | undefined {
	const source = getSourceById(id);
	if (!source) return undefined;

	return {
		id: source.id,
		name: source.name,
		category: source.category,
		active: source.active,
		priority: source.priority,
	};
}

/**
 * Get all unique categories
 */
export function getCategories(): string[] {
	const categories = new Set(
		DEFAULT_NEWS_SOURCES
			.filter((source) => source.active)
			.map((source) => source.category)
	);
	return Array.from(categories).sort();
}

/**
 * Check if a source is active
 */
export function isSourceActive(id: string): boolean {
	const source = getSourceById(id);
	return source?.active ?? false;
}

/**
 * Get total number of sources
 */
export function getSourceCount(): number {
	return DEFAULT_NEWS_SOURCES.length;
}

/**
 * Get number of active sources
 */
export function getActiveSourceCount(): number {
	return DEFAULT_NEWS_SOURCES.filter((source) => source.active).length;
}
