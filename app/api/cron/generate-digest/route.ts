import { NextRequest, NextResponse } from 'next/server';
import { getStorage } from '@/lib/storage';
import { generateDailyDigest } from '@/lib/feed/digest-generator';
import { logger } from '@/lib/logger';

export async function GET(request: NextRequest) {
  // Verify Vercel Cron secret token for security
  const authHeader = request.headers.get('authorization');
  if (authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
    logger.warn('Unauthorized cron request');
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    logger.info('Starting digest generation');
    const storage = getStorage();
    const digest = await generateDailyDigest(storage);

    logger.info('Digest generated successfully', {
      date: digest.date,
      articleCount: digest.articles.length,
    });

    return NextResponse.json({
      success: true,
      date: digest.date,
      articleCount: digest.articles.length,
      message: 'Digest generated successfully',
    });
  } catch (error) {
    logger.error('Cron job error', error instanceof Error ? error : new Error(String(error)));
    return NextResponse.json(
      {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      },
      { status: 500 }
    );
  }
}

// Set timeout to 5 minutes (digest generation takes ~70-75 seconds)
export const maxDuration = 300;
