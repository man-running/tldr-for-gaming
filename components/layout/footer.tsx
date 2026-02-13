"use client";

import { EmailSubscriptionForm } from "@takara/shared-ui/email-subscription-form";
import { FooterContainer } from "@takara/shared-ui/footer-container";
import { FooterShader } from "@takara/shared-ui/footer-shader";
import { Rss } from "lucide-react";
import Link from "next/link";
import { siteConfig } from "@/lib/constants";

export function Footer() {
	const currentYear = new Date().getFullYear();

	return (
		<FooterContainer>
			<div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6 mb-20 border-b border-dashed pb-20">
				<div className="flex flex-col gap-1">
					<h3 className="text-lg font-semibold">Stay in the loop</h3>
					<p className="text-sm opacity-50">
						Get tldr.takara.ai to Your Email, Everyday.
					</p>
				</div>
				<EmailSubscriptionForm
					endpoint="/api/subscribe"
					turnstileSiteKey={process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY}
				/>
			</div>
			<div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
				<div className="flex flex-wrap items-center gap-2 text-[15px] leading-[191%] tracking-[0.025em]">
					<Link href="/" className="group flex items-center">
						<span className="font-black text-foreground">tldr.</span>
						<span className="font-black text-accent">takara.ai</span>
						<span className="sr-only">Home</span>
					</Link>
					<span className="text-muted-foreground">·</span>
					<span className="text-sm text-muted-foreground sm:hidden">
						Daily at 6am UTC
					</span>
					<span className="hidden sm:inline text-sm text-muted-foreground">
						{siteConfig.description}
					</span>
					<span className="text-muted-foreground">·</span>
					<span className="text-sm text-muted-foreground">
						© {currentYear} takara.ai Ltd
					</span>
				</div>
				<nav className="flex flex-wrap items-center gap-4">
					<Link
						// @ts-expect-error External or non-app route not part of typed routes
						href={siteConfig.links.rssFeed}
						target="_blank"
						rel="noopener noreferrer"
						className="link-standard"
						aria-label="RSS Feed"
						data-ph-capture-attribute-action="rss-subscribe"
					>
						<Rss className="h-4 w-4 text-accent" aria-hidden="true" />
						<span>RSS</span>
					</Link>
				</nav>
			</div>
			<p className="mt-4 text-xs text-muted-foreground">
				Content is sourced from third-party publications.
			</p>
			<FooterShader image="/takara.svg" />
		</FooterContainer>
	);
}
