import { NextRequest, NextResponse } from "next/server";
import { getStorage } from "@/lib/storage";

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
        return NextResponse.json(
          { error: "Invalid date format. Use YYYY-MM-DD" },
          { status: 400 }
        );
      }
      targetDate = parsed.toISOString().split("T")[0];
    } else {
      targetDate = new Date().toISOString().split("T")[0];
    }

    // Try to get digest for the specified date
    let digest = await storage.getDigest(targetDate);

    // If no digest for today, try to get the latest digest
    if (!digest && !dateParam) {
      console.log(`No digest found for ${targetDate}, trying latest...`);
      digest = await storage.getLatestDigest();
    }

    if (!digest) {
      return NextResponse.json(
        {
          error: "No digest available",
          message: "Run 'npm run generate-digest' to create a digest",
        },
        { status: 404 }
      );
    }

    return NextResponse.json(digest, {
      headers: {
        "Content-Type": "application/json",
        "Cache-Control": "public, s-maxage=3600, stale-while-revalidate=86400",
      },
    });
  } catch (error) {
    console.error("Error retrieving digest:", error);
    return NextResponse.json(
      { error: "Failed to retrieve digest" },
      { status: 500 }
    );
  }
}
