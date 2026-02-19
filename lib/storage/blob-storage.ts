/**
 * Vercel Blob Storage implementation
 * Stores data as JSON blobs in Vercel Blob Storage
 */

import { put, list, del } from '@vercel/blob';
import type {
  IStorage,
  Article,
  StoredSummary,
  DailyDigest,
} from './storage-interface';

export class BlobStorage implements IStorage {
  /**
   * Blob path structure mirrors file system:
   * - articles/{date}.json
   * - summaries/{id}.json
   * - digests/{date}.json
   */

  /**
   * Initialize storage (no-op for blob storage)
   */
  async initialize(): Promise<void> {
    // No initialization needed for blob storage
  }

  /**
   * Save articles for a specific date
   */
  async saveArticles(date: string, articles: Article[]): Promise<void> {
    const pathname = `articles/${date}.json`;
    await put(pathname, JSON.stringify(articles, null, 2), {
      access: 'public',
      contentType: 'application/json',
      allowOverwrite: true,
    });
    console.log(`Saved ${articles.length} articles to blob: ${pathname}`);
  }

  /**
   * Get articles for a specific date
   */
  async getArticles(date: string): Promise<Article[] | null> {
    const pathname = `articles/${date}.json`;

    try {
      // List blobs to find the specific file
      const { blobs } = await list({ prefix: pathname, limit: 1 });

      if (blobs.length === 0) {
        return null;
      }

      const response = await fetch(blobs[0].url);
      if (!response.ok) {
        return null;
      }

      const data = await response.text();
      return JSON.parse(data);
    } catch (error) {
      console.error(`Error fetching articles for ${date}:`, error);
      return null;
    }
  }

  /**
   * Save a single article summary
   */
  async saveSummary(summary: StoredSummary): Promise<void> {
    const pathname = `summaries/${summary.id}.json`;
    await put(pathname, JSON.stringify(summary, null, 2), {
      access: 'public',
      contentType: 'application/json',
      allowOverwrite: true,
    });
    console.log(`Saved summary to blob: ${summary.id}`);
  }

  /**
   * Get a summary by ID
   */
  async getSummary(id: string): Promise<StoredSummary | null> {
    const pathname = `summaries/${id}.json`;

    try {
      // List blobs to find the specific file
      const { blobs } = await list({ prefix: pathname, limit: 1 });

      if (blobs.length === 0) {
        return null;
      }

      const response = await fetch(blobs[0].url);
      if (!response.ok) {
        return null;
      }

      const data = await response.text();
      return JSON.parse(data);
    } catch (error) {
      console.error(`Error fetching summary ${id}:`, error);
      return null;
    }
  }

  /**
   * Get all summaries
   */
  async getAllSummaries(): Promise<StoredSummary[]> {
    try {
      const { blobs } = await list({ prefix: 'summaries/' });
      const summaries: StoredSummary[] = [];

      for (const blob of blobs) {
        try {
          const response = await fetch(blob.url);
          if (response.ok) {
            const data = await response.text();
            summaries.push(JSON.parse(data));
          }
        } catch (error) {
          console.error(`Error fetching summary from ${blob.url}:`, error);
        }
      }

      return summaries;
    } catch (error) {
      console.error('Error listing summaries:', error);
      return [];
    }
  }

  /**
   * Save a daily digest
   */
  async saveDigest(digest: DailyDigest): Promise<void> {
    const pathname = `digests/${digest.date}.json`;
    await put(pathname, JSON.stringify(digest, null, 2), {
      access: 'public',
      contentType: 'application/json',
      allowOverwrite: true,
    });
    console.log(`Saved digest for ${digest.date} to blob: ${pathname}`);
  }

  /**
   * Get digest for a specific date
   */
  async getDigest(date: string): Promise<DailyDigest | null> {
    const pathname = `digests/${date}.json`;

    try {
      // List blobs to find the specific file
      const { blobs } = await list({ prefix: pathname, limit: 1 });

      if (blobs.length === 0) {
        return null;
      }

      const response = await fetch(blobs[0].url);
      if (!response.ok) {
        return null;
      }

      const data = await response.text();
      return JSON.parse(data);
    } catch (error) {
      console.error(`Error fetching digest for ${date}:`, error);
      return null;
    }
  }

  /**
   * Get the most recent digest
   */
  async getLatestDigest(): Promise<DailyDigest | null> {
    try {
      const { blobs } = await list({ prefix: 'digests/' });

      if (blobs.length === 0) {
        return null;
      }

      // Sort by pathname (which contains ISO date) in descending order
      const sortedBlobs = blobs
        .sort((a, b) => b.pathname.localeCompare(a.pathname));

      const latestBlob = sortedBlobs[0];
      const response = await fetch(latestBlob.url);

      if (!response.ok) {
        return null;
      }

      const data = await response.text();
      return JSON.parse(data);
    } catch (error) {
      console.error('Error fetching latest digest:', error);
      return null;
    }
  }
}
