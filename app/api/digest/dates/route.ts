import { NextResponse } from "next/server";
import { getStorage } from "@/lib/storage";
import { logger } from "@/lib/logger";

const storage = getStorage();

export async function GET() {
  try {
    await storage.initialize();
    const dates = await storage.listDigestDates();
    return NextResponse.json({ dates }, {
      headers: {
        "Cache-Control": "public, s-maxage=3600, stale-while-revalidate=86400",
      },
    });
  } catch (error) {
    logger.error("Error listing digest dates", error instanceof Error ? error : new Error(String(error)));
    return NextResponse.json({ error: "Failed to list digest dates" }, { status: 500 });
  }
}
