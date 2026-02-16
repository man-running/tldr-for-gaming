import { NextRequest, NextResponse } from 'next/server';
import { getStorage } from '@/lib/storage';

const storage = getStorage();

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    // Initialize storage
    await storage.initialize();

    // Get the ID from params
    const { id } = await params;

    // Get summary from storage
    const summary = await storage.getSummary(id);

    if (!summary) {
      return NextResponse.json(
        {
          error: 'Summary not found',
          message: 'This summary may not have been generated yet. Run "npm run generate-digest" to create summaries.',
        },
        { status: 404 }
      );
    }

    return NextResponse.json(summary, {
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'public, s-maxage=3600, stale-while-revalidate=86400',
      },
    });
  } catch (error) {
    console.error('Error retrieving summary:', error);
    return NextResponse.json(
      { error: 'Failed to retrieve summary' },
      { status: 500 }
    );
  }
}
