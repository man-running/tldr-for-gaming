import type { Metadata } from "next";
import { FeedContent } from "@/components/feed/feed-content";
import { FeedDataProvider } from "@/components/feed/feed-data-provider";
import { HeaderContent } from "@/components/feed/header-content";
import { RestoreScroll } from "@/components/feed/restore-scroll";
import { Footer } from "@/components/layout/footer";
import { HomePageNavbarProvider } from "@/components/layout/home-page-navbar-provider";
import { Navbar } from "@/components/layout/navbar";
import { NavbarSection } from "@/components/layout/navbar-section";
import { Section } from "@/components/layout/section";

export const dynamic = "force-static";

export function generateMetadata(): Metadata {
	return {
		title: "TLDR - Daily AI Research Summaries",
		description:
			"Daily summaries of the latest AI research papers from ArXiv and HuggingFace",
		openGraph: {
			title: "TLDR - Daily AI Research Summaries",
			description:
				"Daily summaries of the latest AI research papers from ArXiv and HuggingFace",
			url: "https://tldr.takara.ai",
			type: "website",
			images: ["/OG-Image-TLDR.jpg"],
		},
		twitter: {
			card: "summary_large_image",
		},
	};
}

export default function Home() {
	return (
		<FeedDataProvider>
			<HomePageNavbarProvider>
				<NavbarSection>
					<Navbar />
				</NavbarSection>
				<RestoreScroll />
				<main className="min-h-screen w-full flex flex-col mb-40">
					<HeaderContent />

					<Section className="w-full items-stretch">
						<FeedContent />
					</Section>
				</main>
				<Footer />
			</HomePageNavbarProvider>
		</FeedDataProvider>
	);
}
