/**
 * Storage Factory
 * Environment-aware storage selection
 */

import { FileStorage } from './file-storage';
import { BlobStorage } from './blob-storage';
import type { IStorage } from './storage-interface';

/**
 * Get storage implementation based on environment
 * - Uses Blob Storage in Vercel environment
 * - Uses File Storage for local development
 */
export function getStorage(): IStorage {
  // Use Blob Storage in Vercel environment
  if (process.env.VERCEL || process.env.BLOB_READ_WRITE_TOKEN) {
    return new BlobStorage();
  }

  // Use File Storage for local development
  return new FileStorage('./data');
}

// Re-export types and interfaces for convenience
export * from './storage-interface';
export { FileStorage } from './file-storage';
export { BlobStorage } from './blob-storage';
