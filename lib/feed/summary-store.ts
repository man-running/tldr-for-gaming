/**
 * In-memory summary store for article summaries
 * Stores Claude-generated summaries keyed by article ID
 */

export interface StoredSummary {
  id: string;
  title: string;
  summary: string;
  url: string;
  sourceName: string;
  publishedDate: string;
  generatedAt: string;
}

// In-memory storage
const summaries = new Map<string, StoredSummary>();

/**
 * Store a summary
 */
export function storeSummary(summary: StoredSummary): void {
  console.log(`[STORE] Storing summary with ID: ${summary.id}`);
  summaries.set(summary.id, summary);
  console.log(`[STORE] Total summaries stored: ${summaries.size}`);
}

/**
 * Retrieve a summary by ID
 */
export function getSummary(id: string): StoredSummary | undefined {
  console.log(`[STORE] Looking up summary with ID: ${id}`);
  console.log(`[STORE] Available IDs: ${Array.from(summaries.keys()).join(', ')}`);
  const result = summaries.get(id);
  console.log(`[STORE] Found: ${result ? 'YES' : 'NO'}`);
  return result;
}

/**
 * Get all summaries
 */
export function getAllSummaries(): StoredSummary[] {
  return Array.from(summaries.values());
}

/**
 * Clear all summaries (useful for testing)
 */
export function clearSummaries(): void {
  summaries.clear();
}

/**
 * Check if a summary exists
 */
export function hasSummary(id: string): boolean {
  return summaries.has(id);
}
