export interface FeedItemType {
	title: string;
	link: string;
	description: string;
	pubDate: string;
	guid: string;
}

export interface RssFeed {
	title: string;
	description: string;
	link: string;
	lastBuildDate?: string;
	items: FeedItemType[];
	blobURL?: string; // Optional: URL for client to fetch directly
}
