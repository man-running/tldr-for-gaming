/**
 * File-based storage implementation
 * Stores data as JSON files in the data/ directory
 */

import * as fs from 'fs';
import * as path from 'path';
import type {
  IStorage,
  Article,
  StoredSummary,
  DailyDigest,
} from './storage-interface';

export class FileStorage implements IStorage {
  private dataDir: string;
  private articlesDir: string;
  private summariesDir: string;
  private digestsDir: string;

  constructor(dataDir: string = './data') {
    this.dataDir = dataDir;
    this.articlesDir = path.join(dataDir, 'articles');
    this.summariesDir = path.join(dataDir, 'summaries');
    this.digestsDir = path.join(dataDir, 'digests');
  }

  /**
   * Initialize storage directories
   */
  async initialize(): Promise<void> {
    const dirs = [this.dataDir, this.articlesDir, this.summariesDir, this.digestsDir];

    for (const dir of dirs) {
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
        console.log(`Created directory: ${dir}`);
      }
    }
  }

  /**
   * Save articles for a specific date
   */
  async saveArticles(date: string, articles: Article[]): Promise<void> {
    const filePath = path.join(this.articlesDir, `${date}.json`);
    fs.writeFileSync(filePath, JSON.stringify(articles, null, 2), 'utf-8');
    console.log(`Saved ${articles.length} articles to ${filePath}`);
  }

  /**
   * Get articles for a specific date
   */
  async getArticles(date: string): Promise<Article[] | null> {
    const filePath = path.join(this.articlesDir, `${date}.json`);

    if (!fs.existsSync(filePath)) {
      return null;
    }

    const data = fs.readFileSync(filePath, 'utf-8');
    return JSON.parse(data);
  }

  /**
   * Save a single article summary
   */
  async saveSummary(summary: StoredSummary): Promise<void> {
    const filePath = path.join(this.summariesDir, `${summary.id}.json`);
    fs.writeFileSync(filePath, JSON.stringify(summary, null, 2), 'utf-8');
    console.log(`Saved summary: ${summary.id}`);
  }

  /**
   * Get a summary by ID
   */
  async getSummary(id: string): Promise<StoredSummary | null> {
    const filePath = path.join(this.summariesDir, `${id}.json`);

    if (!fs.existsSync(filePath)) {
      return null;
    }

    const data = fs.readFileSync(filePath, 'utf-8');
    return JSON.parse(data);
  }

  /**
   * Get all summaries
   */
  async getAllSummaries(): Promise<StoredSummary[]> {
    if (!fs.existsSync(this.summariesDir)) {
      return [];
    }

    const files = fs.readdirSync(this.summariesDir);
    const summaries: StoredSummary[] = [];

    for (const file of files) {
      if (file.endsWith('.json')) {
        const filePath = path.join(this.summariesDir, file);
        const data = fs.readFileSync(filePath, 'utf-8');
        summaries.push(JSON.parse(data));
      }
    }

    return summaries;
  }

  /**
   * Save a daily digest
   */
  async saveDigest(digest: DailyDigest): Promise<void> {
    const filePath = path.join(this.digestsDir, `${digest.date}.json`);
    fs.writeFileSync(filePath, JSON.stringify(digest, null, 2), 'utf-8');
    console.log(`Saved digest for ${digest.date} to ${filePath}`);
  }

  /**
   * Get digest for a specific date
   */
  async getDigest(date: string): Promise<DailyDigest | null> {
    const filePath = path.join(this.digestsDir, `${date}.json`);

    if (!fs.existsSync(filePath)) {
      return null;
    }

    const data = fs.readFileSync(filePath, 'utf-8');
    return JSON.parse(data);
  }

  /**
   * Get the most recent digest
   */
  async getLatestDigest(): Promise<DailyDigest | null> {
    if (!fs.existsSync(this.digestsDir)) {
      return null;
    }

    const files = fs.readdirSync(this.digestsDir)
      .filter(f => f.endsWith('.json'))
      .sort()
      .reverse();

    if (files.length === 0) {
      return null;
    }

    const latestFile = files[0];
    const filePath = path.join(this.digestsDir, latestFile);
    const data = fs.readFileSync(filePath, 'utf-8');
    return JSON.parse(data);
  }
}
