export const siteConfig = {
	name: "iGaming TLDR",
	description:
		"Daily curated iGaming and sports betting news summaries. Top 5 stories every morning.",
	url: process.env.NEXT_PUBLIC_SITE_URL || "https://gaming-tldr.example.com",
	ogImage: "/OG-Image-TLDR.jpg",
	links: {
		sourceIgamingBusiness: "https://www.igamingbusiness.com",
		sourceGamblingInsider: "https://www.gamblinginsider.com",
		sourceEgamingReview: "https://www.egamingreview.com",
		sourceSourcetech: "https://www.sportech.com",
		sourceBeIndustry: "https://www.bettingindustry.com",
		rssFeed: "/api/news-feed",
	},
	creator: "iGaming TLDR",
};

const _feedConfig = {
	url: "/api/news-tldr",
	cacheTime: 5 * 60 * 1000, // 5 minutes
	retryAttempts: 3,
	retryDelay: 1000, // 1 second
	updateInterval: 4 * 60 * 60 * 1000, // 4 hours (fetch fresh news regularly)
	dailyUpdateTime: "07:00", // UTC time for daily top 5
};

// Article sourcing configuration
const _articleConfig = {
	topArticlesPerDay: 5,
	minArticlesFromEachSource: 0, // 0 means sources can contribute any number
	diversifyByCategory: true,
	summaryMaxLength: 300, // characters
	// Ranking criteria weights (must sum to 1.0)
	rankingWeights: {
		recency: 0.40,      // 40% - Recent articles score higher
		source: 0.30,       // 30% - Trusted sources weight
		engagement: 0.20,   // 20% - Comments, shares (if available)
		diversity: 0.10,    // 10% - Category diversity factor
	},
};

// Debug configuration
const _debugConfig = {
	rssLogging:
		process.env.NODE_ENV === "development" && process.env.RSS_DEBUG === "true",
	articleLogging:
		process.env.NODE_ENV === "development" && process.env.ARTICLE_DEBUG === "true",
};

const _HEADLINE_TEXT = "Daily iGaming Digest";

// Article categories for iGaming content
export const ARTICLE_CATEGORIES = {
	REGULATIONS: "Regulations",
	BUSINESS: "Business",
	TECHNOLOGY: "Technology",
	SPORTS_BETTING: "Sports Betting",
	M_AND_A: "M&A",
	INTERNATIONAL: "International",
	PAYMENTS: "Payments",
	RESPONSIBLE_GAMING: "Responsible Gaming",
} as const;

// Export configurations
export const feedConfig = _feedConfig;
export const articleConfig = _articleConfig;
export const debugConfig = _debugConfig;
export const HEADLINE_TEXT = _HEADLINE_TEXT;
