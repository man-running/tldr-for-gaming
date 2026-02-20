import type { Metadata } from "next";
import { DigestDisplay } from "@/components/feed/digest-display";
import { Navbar } from "@/components/layout/navbar";
import { NavbarSection } from "@/components/layout/navbar-section";
import { Footer } from "@/components/layout/footer";
import { Section } from "@/components/layout/section";

interface DigestDatePageProps {
  params: Promise<{
    date: string;
  }>;
}

export async function generateMetadata({
  params,
}: DigestDatePageProps): Promise<Metadata> {
  const { date } = await params;
  return {
    title: `Daily Digest - ${date} - TLDR Gaming`,
    description: `Daily digest of top iGaming news articles for ${date}, ranked and summarized with AI.`,
  };
}

export default async function DigestDatePage({ params }: DigestDatePageProps) {
  const { date } = await params;

  // Validate date format
  if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) {
    return (
      <>
        <NavbarSection>
          <Navbar />
        </NavbarSection>
        <main className="min-h-screen w-full flex flex-col mb-40">
          <Section className="w-full items-stretch">
            <div className="w-full max-w-4xl mx-auto">
              <div className="rounded-lg border border-red-200 bg-red-50 p-6 dark:border-red-800 dark:bg-red-900/20">
                <h1 className="font-semibold text-red-900 dark:text-red-200">
                  Invalid Date Format
                </h1>
                <p className="text-sm text-red-800 dark:text-red-300 mt-2">
                  Please use the format YYYY-MM-DD in the URL.
                </p>
              </div>
            </div>
          </Section>
        </main>
        <Footer />
      </>
    );
  }

  return (
    <>
      <NavbarSection>
        <Navbar />
      </NavbarSection>
      <main className="min-h-screen w-full flex flex-col mb-40">
        <Section className="w-full items-stretch">
          <div className="w-full max-w-4xl mx-auto">
            <DigestDisplay date={date} />
          </div>
        </Section>
      </main>
      <Footer />
    </>
  );
}
