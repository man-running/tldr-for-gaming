import type { Metadata } from "next";
import { Footer } from "@/components/layout/footer";
import { PaperContent } from "@/components/paper/paper-content";
import { PaperDataProvider } from "@/components/paper/paper-data-provider";
import {
	constructSpectrogramBlobURL,
	fetchImageWithFallback,
} from "@/lib/image-utils";

export const revalidate = 604800; // ISR: revalidate cached pages every week

export async function generateStaticParams() {
	// Generate empty array - pages will be generated on-demand and then cached
	return [];
}

export async function generateMetadata({
	params,
}: {
	params: Promise<{ arxivId: string }>;
}): Promise<Metadata> {
	const { arxivId } = await params;

	try {
		const baseUrl = process.env.BASE_URL || "https://tldr.takara.ai";

		// Fetch paper data for metadata
		const paperResponse = await fetch(`${baseUrl}/api/paper?id=${arxivId}`, {
			next: { revalidate: 86400 }, // Cache for 24 hours
		}).then(async (res) => {
			const json = await res.json();
			// If blob URL is provided, fetch paper data directly from blob
			if (json?.success && json?.blobURL && !json?.data) {
				try {
					const blobResponse = await fetch(json.blobURL, {
						next: { revalidate: 86400 },
					});
					if (blobResponse.ok) {
						const blobData = await blobResponse.json();
						return { ...json, data: blobData };
					}
				} catch (err) {
					// Fallback to original response if blob fetch fails
					console.error("Failed to fetch from blob URL:", err);
				}
			}
			return json;
		});

		if (!paperResponse?.success || !paperResponse?.data) {
			return {
				title: "Paper Not Found - Takara TLDR",
				description: "The requested research paper could not be found.",
			};
		}

		const paper = paperResponse.data;

		// Generate dynamic OG image URL
		const ogImageUrl = `${baseUrl}/api/og?id=${arxivId}`;

		// Truncate abstract for description
		const description = paper.abstract
			? paper.abstract.length > 160
				? `${paper.abstract.substring(0, 157)}...`
				: paper.abstract
			: "AI research paper summary and details";

		// Format authors for metadata
		const authors = paper.authors?.slice(0, 3).join(", ") || "";
		const authorText = authors ? ` by ${authors}` : "";

		return {
			title: `${paper.title} - Takara TLDR`,
			description: description,
			authors: paper.authors?.map((author: string) => ({ name: author })),
			openGraph: {
				type: "article",
				title: paper.title,
				description: description,
				images: [
					{
						url: ogImageUrl,
						width: 1200,
						height: 630,
						alt: `${paper.title}${authorText}`,
					},
				],
			},
			twitter: {
				card: "summary_large_image",
				title: paper.title,
				description: description,
				images: [ogImageUrl],
			},
		};
	} catch (_error) {
		// Fallback to static metadata on error
		return {
			title: "Research Paper - Takara TLDR",
			description: "AI research paper summary and details",
			openGraph: {
				type: "article",
				images: [
					{
						url: "/OG-Image-TLDR.jpg",
						width: 2400,
						height: 1256,
					},
				],
			},
			twitter: {
				card: "summary_large_image",
				images: ["/OG-Image-TLDR.jpg"],
			},
		};
	}
}

export default async function PaperPage({
	params,
}: {
	params: Promise<{ arxivId: string }>;
}) {
	const { arxivId } = await params;

	// Server-side fetch for ISR - cache paper data for 24 hours since it's static
	const baseUrl = process.env.BASE_URL || "https://tldr.takara.ai";

	// First fetch the paper data
	const paperResponse = await fetch(`${baseUrl}/api/paper?id=${arxivId}`, {
		next: { revalidate: 86400 }, // Cache paper for 24 hours - static content
	})
		.then(async (res) => {
			const json = await res.json();
			// If blob URL is provided, fetch paper data directly from blob
			if (json?.success && json?.blobURL && !json?.data) {
				try {
					const blobResponse = await fetch(json.blobURL, {
						next: { revalidate: 86400 },
					});
					if (blobResponse.ok) {
						const blobData = await blobResponse.json();
						return { ...json, data: blobData };
					}
				} catch (err) {
					// Fallback to original response if blob fetch fails
					console.error("Failed to fetch from blob URL:", err);
				}
			}
			return json;
		})
		.catch(() => null);

	// Then fetch the featured image using the paper title (spectrogram)
	const featuredImage =
		paperResponse?.success && paperResponse?.data?.title
			? await (async () => {
				try {
					const blobURL = constructSpectrogramBlobURL(
						paperResponse.data.title,
					);
					const apiURL = `${baseUrl}/api/spectrogram?q=${encodeURIComponent(paperResponse.data.title)}&format=json`;

					const imageURL = await fetchImageWithFallback(blobURL, apiURL);

					return {
						url: imageURL,
						alt: `DS1 spectrogram: ${paperResponse.data.title}`,
					};
				} catch (error) {
					console.error("Failed to fetch featured image:", error);
					return null;
				}
			})()
			: null;

	// If paper doesn't exist, return 404
	if (!paperResponse?.success || !paperResponse?.data) {
		throw new Error("Paper not found");
	}

	return (
		<>
			<PaperDataProvider
				initialPaper={paperResponse.data}
				initialFeaturedImage={featuredImage}
			>
				<PaperContent arxivId={arxivId} />
			</PaperDataProvider>
			<Footer />
		</>
	);
}
