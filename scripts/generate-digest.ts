#!/usr/bin/env node
/**
 * Background job to generate daily digest
 * Run this via: npm run generate-digest
 * Or set up as a cron job to run daily
 */

import { config } from 'dotenv';
import { resolve } from 'path';

// Load environment variables from .env.local
config({ path: resolve(process.cwd(), '.env.local') });

import { getStorage } from '../lib/storage';
import { generateDailyDigest } from '../lib/feed/digest-generator';

async function main() {
  try {
    const storage = getStorage();
    const digest = await generateDailyDigest(storage);

    console.log(`\nDigest saved to: data/digests/${digest.date}.json`);
    console.log('\nYou can now view the digest at: http://localhost:3001/digest\n');
  } catch (error) {
    console.error('\n❌ Error generating digest:', error);
    process.exit(1);
  }
}

main();
