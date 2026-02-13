import { Ray } from "@takara/shared-ui";
import { MathJaxContext } from "better-react-mathjax";
import type { Metadata } from "next";
import { Lato } from "next/font/google";
import type React from "react";
import { PostHogProvider } from "@/components/analytics/posthog-provider";
import { RootNavbar } from "@/components/layout/root-navbar";
import { ThemeProvider } from "@/components/ui/theme-provider";
import { siteConfig } from "@/lib/constants";
import "./globals.css";

// Font configuration
const lato = Lato({
	subsets: ["latin"],
	weight: ["300", "400", "700", "900"],
	variable: "--font-lato", // Optional: Define CSS variable
});

export const viewport = {
	width: "device-width",
	initialScale: 1,
	maximumScale: 5,
};

export const metadata: Metadata = {
	title: {
		default: siteConfig.name,
		template: `%s | ${siteConfig.name}`,
	},
	description: siteConfig.description,
	applicationName: siteConfig.name,
	authors: [
		{ name: siteConfig.creator, url: siteConfig.links.takaraHuggingFace },
	],
	creator: siteConfig.creator,
	publisher: "takara.ai Ltd",
	keywords: [
		"AI",
		"research",
		"papers",
		"summaries",
		"takara.ai",
		"machine learning",
		"artificial intelligence",
	],
	metadataBase: new URL(siteConfig.url),
	alternates: {
		types: {
			"application/rss+xml": [
				{ url: siteConfig.links.rssFeed, title: "Takara TLDR RSS Feed" },
			],
		},
	},
	openGraph: {
		type: "website",
		locale: "en_US",
		url: siteConfig.url,
		title: siteConfig.name,
		description: siteConfig.description,
		siteName: siteConfig.name,
		images: [
			{
				url: siteConfig.ogImage,
				width: 2400,
				height: 1256,
				alt: siteConfig.name,
			},
		],
	},
	twitter: {
		card: "summary_large_image",
		title: siteConfig.name,
		description: siteConfig.description,
		images: [siteConfig.ogImage],
		creator: "@takara_ai",
	},
	robots: {
		index: true,
		follow: true,
		googleBot: {
			index: true,
			follow: true,
			"max-image-preview": "large",
			"max-snippet": -1,
		},
	},
	icons: {
		icon: [
			{ url: "/assets/og/icon.svg", type: "image/svg+xml" },
			{ url: "/favicon.ico", sizes: "any" },
		],
		shortcut: "/favicon-16x16.png",
		apple: "/apple-icon.png",
	},
};

export default function RootLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	return (
		<html lang="en" suppressHydrationWarning>
			<head />
			<body className={`${lato.variable} font-lato antialiased relative`}>
				<PostHogProvider>
					<ThemeProvider attribute="class" defaultTheme="system" enableSystem>
						<MathJaxContext version={3}>
							<RootNavbar />
							<main>{children}</main>
							<div className="pointer-events-none absolute inset-0 overflow-hidden -z-10">
								<Ray className="absolute left-2/3 top-0 size-300 text-accent fill-accent blur-2xl rotate-45 -translate-x-1/2 -translate-y-1/2 opacity-20" />
								<Ray className="absolute left-3/4 top-100 size-400 text-accent fill-accent blur-2xl rotate-45 -translate-x-1/2 -translate-y-1/2 opacity-10" />
								<Ray className="absolute left-0 top-140 size-300 text-accent fill-accent blur-2xl rotate-45 -translate-x-1/2 -translate-y-1/2 opacity-20" />
							</div>
						</MathJaxContext>
					</ThemeProvider>
				</PostHogProvider>
			</body>
		</html>
	);
}
