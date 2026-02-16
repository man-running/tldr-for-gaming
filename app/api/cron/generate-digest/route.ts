import { NextRequest, NextResponse } from 'next/server';
import { getStorage } from '@/lib/storage';
import { generateDailyDigest } from '@/lib/feed/digest-generator';

export async function GET(request: NextRequest) {
  // Verify Vercel Cron secret token for security
  const authHeader = request.headers.get('authorization');
  if (authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const storage = getStorage();
    const digest = await generateDailyDigest(storage);

    return NextResponse.json({
      success: true,
      date: digest.date,
      articleCount: digest.articles.length,
      message: 'Digest generated successfully',
    });
  } catch (error) {
    console.error('Cron job error:', error);
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
