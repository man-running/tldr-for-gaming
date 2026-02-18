import { NextRequest, NextResponse } from "next/server";
import { getStorage } from "@/lib/storage";
import { logger } from "@/lib/logger";

const storage = getStorage();

export async function GET(request: NextRequest) {
  try {
    // Initialize storage (creates directories if needed)
    await storage.initialize();

    // Get date parameter from query string
    const searchParams = request.nextUrl.searchParams;
    const dateParam = searchParams.get("date");

    // Default to today's date
    let targetDate: string;
    if (dateParam) {
      const parsed = new Date(dateParam);
      if (isNaN(parsed.getTime())) {
        logger.warn("Invalid date format in request", { dateParam });
        return NextResponse.json(
          { error: "Invalid date format. Use YYYY-MM-DD" },
          { status: 400 }
        );
      }
      targetDate = parsed.toISOString().split("T")[0];
    } else {
      targetDate = new Date().toISOString().split("T")[0];
    }

    logger.debug("Retrieving digest", { targetDate });

    // Try to get digest for the specified date
    let digest = await storage.getDigest(targetDate);

    // If no digest for today, try to get the latest digest
    if (!digest && !dateParam) {
      logger.info("No digest found for requested date, fetching latest", { targetDate });
      digest = await storage.getLatestDigest();
    }

    if (!digest) {
      logger.warn("No digest available");
      return NextResponse.json(
        {
          error: "No digest available",
          message: "Run 'npm run generate-digest' to create a digest",
        },
        { status: 404 }
      );
    }

    logger.info("Digest retrieved successfully", { date: targetDate });

    return NextResponse.json(digest, {
      headers: {
        "Content-Type": "application/json",
        "Cache-Control": "public, s-maxage=3600, stale-while-revalidate=86400",
      },
    });
  } catch (error) {
    logger.error("Error retrieving digest", error instanceof Error ? error : new Error(String(error)));
    return NextResponse.json(
      { error: "Failed to retrieve digest" },
      { status: 500 }
    );
  }
}
