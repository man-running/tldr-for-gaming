export const siteConfig = {
	name: "Takara TLDR",
	description:
		"Catch the latest AI research daily at 6am UTC, summarised by takara.ai",
	url: "https://tldr.takara.ai",
	ogImage: "/OG-Image-TLDR.jpg",
	links: {
		takaraWebsite: "https://takara.ai",
		takaraHuggingFace: "https://huggingface.co/takara-ai",
		rssFeed: "https://papers.takara.ai/api/tldr",
	},
	creator: "takara.ai Frontier Research Team",
};

const _feedConfig = {
	url: "https://papers.takara.ai/api/tldr",
	cacheTime: 5 * 60 * 1000, // 5 minutes
	retryAttempts: 3,
	retryDelay: 1000, // 1 second
};

// Debug configuration
const _debugConfig = {
	rssLogging:
		process.env.NODE_ENV === "development" && process.env.RSS_DEBUG === "true",
};

const _HEADLINE_TEXT = "Morning Headline";
