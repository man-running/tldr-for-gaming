/**
 * Storage Interface - Abstract storage operations
 * Allows easy migration from file system → SQLite → PostgreSQL
 */

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

export interface StoredSummary {
  id: string;
  title: string;
  summary: string;
  url: string;
  sourceName: string;
  publishedDate: string;
  generatedAt: string;
  imageUrl?: string;
}

export interface RankedArticle {
  article: Article;
  score: number;
  rank: number;
  reason: string;
}

export interface DailyDigest {
  date: string;
  articles: RankedArticle[];
  headline: string;
  summary: string;
  created: string;
}

/**
 * Storage interface - implement this for different backends
 */
export interface IStorage {
  // Article operations
  saveArticles(date: string, articles: Article[]): Promise<void>;
  getArticles(date: string): Promise<Article[] | null>;

  // Summary operations
  saveSummary(summary: StoredSummary): Promise<void>;
  getSummary(id: string): Promise<StoredSummary | null>;
  getAllSummaries(): Promise<StoredSummary[]>;

  // Digest operations
  saveDigest(digest: DailyDigest): Promise<void>;
  getDigest(date: string): Promise<DailyDigest | null>;
  getLatestDigest(): Promise<DailyDigest | null>;

  // Utility
  initialize(): Promise<void>;
}
