import type { Metadata } from "next";
import { DigestDisplay } from "@/components/feed/digest-display";
import { Footer } from "@/components/layout/footer";

export const metadata: Metadata = {
  title: "Daily Digest - TLDR Gaming",
  description: "Daily digest of top iGaming news articles, ranked and summarized with AI.",
};

export default function DigestPage() {
  return (
    <>
      <main className="min-h-screen w-full">
        <div className="py-12 px-6 sm:px-8">
          <DigestDisplay />
        </div>
      </main>
      <Footer />
    </>
  );
}
