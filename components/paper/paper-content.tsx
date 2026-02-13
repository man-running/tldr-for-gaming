"use client";

import { ArrowUpRight, ExternalLink, LucideDot } from "lucide-react";
import Image from "next/image";
import { useEffect, useState } from "react";
import { formatAbstract } from "@/lib/abstract-formatter";
import { cn } from "@/lib/utils";
import { AbstractRenderer } from "./abstract-renderer";
import { FavoriteButton } from "./favorite-button";
import { usePaper } from "./paper-data-provider";

function formatDate(dateString: string) {
	try {
		return new Date(dateString).toLocaleDateString("en-US", {
			year: "numeric",
			month: "long",
			day: "numeric",
		});
	} catch {
		return dateString;
	}
}

function Subtitle({
	children,
	className,
}: {
	children: React.ReactNode;
	className?: string;
}) {
	return (
		<h2
			className={cn(
				"text-base text-muted-foreground font-medium mb-4 border-b border-dashed pb-2 mt-12",
				className,
			)}
		>
			{children}
		</h2>
	);
}

function AuthorList({ authors }: { authors: string[] }) {
	const [isExpanded, setIsExpanded] = useState(false);
	const maxVisibleAuthors = 5;
	const hasMoreAuthors = authors.length > maxVisibleAuthors;
	const visibleAuthors = isExpanded
		? authors
		: authors.slice(0, maxVisibleAuthors);
	const remainingCount = authors.length - maxVisibleAuthors;

	return (
		<div className="flex flex-wrap gap-2">
			{visibleAuthors.map((author, index) => (
				<span key={index} className="text-foreground">
					{author}
					{index < visibleAuthors.length - 1 && ","}
				</span>
			))}
			{hasMoreAuthors && (
				<button
					type="button"
					onClick={() => setIsExpanded(!isExpanded)}
					className="text-muted-foreground hover:text-foreground cursor-pointer"
					aria-label={
						isExpanded
							? "Show fewer authors"
							: `Show all ${authors.length} authors`
					}
				>
					{isExpanded ? "show less" : `and ${remainingCount} more`}
				</button>
			)}
		</div>
	);
}

function PaperSkeleton() {
	return (
		<div className="min-h-screen">
			<main className="w-full py-8 my-8 md:py-12 md:my-12 px-6 relative z-10 mx-auto max-w-5xl">
				<article className="space-y-lg">
					<div className="w-full h-32 sm:h-48 md:h-64 skeleton-lg rounded-lg" />
					<header className="space-y-md">
						<div className="space-y-md">
							<div className="skeleton-heading w-3/4" />
							<div className="flex gap-md">
								<div className="skeleton-heading" />
								<div className="skeleton-heading" />
							</div>
						</div>
						<div className="flex gap-md">
							{[1, 2, 3].map((i) => (
								<div key={i} className="skeleton-text h-10 w-32" />
							))}
						</div>
					</header>
					<section className="space-y-md">
						<div className="skeleton-heading w-24" />
						<div className="space-y-sm">
							{[1, 2, 3, 4].map((i) => (
								<div key={i} className="skeleton-text" />
							))}
						</div>
					</section>
				</article>
			</main>
		</div>
	);
}

export function PaperContent({ arxivId }: { arxivId: string }) {
	const { paper, loading, error, featuredImage } = usePaper();
	const [mathjaxReady, setMathjaxReady] = useState(false);

	useEffect(() => {
		// Load MathJax script from CDN
		const script = document.createElement("script");
		script.src = "https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js";
		script.async = true;
		script.onload = () => {
			setMathjaxReady(true);
			if (window.MathJax) {
				window.MathJax.typeset?.();
			}
		};
		document.head.appendChild(script);

		return () => {
			// Cleanup: remove script tag on unmount if needed
			if (script.parentNode) {
				script.parentNode.removeChild(script);
			}
		};
	}, []);

	useEffect(() => {
		// Re-typeset after content updates
		if (mathjaxReady && window.MathJax && paper?.abstract) {
			window.MathJax.typeset?.();
		}
	}, [paper?.abstract, mathjaxReady]);

	if (loading) return <PaperSkeleton />;

	if (error || !paper) {
		return (
			<div className="min-h-screen grid place-items-center">
				<div className="text-center max-w-md px-6">
					<h1 className="text-4xl lg:text-5xl font-bold leading-tight tracking-tight text-foreground mb-4">
						Paper Not Found
					</h1>
					<p className="text-muted-foreground">
						The requested research paper could not be found.
					</p>
				</div>
			</div>
		);
	}

	return (
		<main className="mx-auto min-h-screen mb-12 w-full px-6 max-w-5xl">
			<article className="flex flex-col gap-8">
				{featuredImage ? (
					<div className="w-full relative overflow-hidden rounded-lg">
						<div className="absolute inset-0">
							<Image
								src={featuredImage.url}
								alt={featuredImage.alt}
								fill
								className="object-cover object-center opacity-50 -hue-rotate-230 "
								priority
								fetchPriority="high"
							/>
						</div>
						<h1 className="relative p-4 sm:p-6 md:p-8 lg:p-10 text-4xl lg:text-5xl font-bold leading-tight tracking-tight drop-shadow-lg">
							{paper.title}
						</h1>
					</div>
				) : (
					<div className="text-4xl lg:text-5xl font-bold leading-tight mx-auto tracking-tight">
						{paper.title}
					</div>
				)}

				{/* Paper Information */}
				<div className="lg:pb-14 lg:pt-4 pb-8">
					<div className="flex items-center gap-2 justify-center text-muted-foreground mb-8">
						{paper.publishedDate && (
							<span>{formatDate(paper.publishedDate)}</span>
						)}

						{paper.publishedDate &&
							paper.authors &&
							paper.authors.length > 0 && <LucideDot className="size-5" />}

						<span>{arxivId}</span>
						<LucideDot className="size-5" />
						<FavoriteButton arxivId={arxivId} />
					</div>

					<div className="flex flex-col gap-6">
						{paper.authors && paper.authors.length > 0 && (
							<div>
								<Subtitle className="mt-0">Authors</Subtitle>
								<AuthorList authors={paper.authors} />
							</div>
						)}
					</div>
				</div>

				{/* Body Content */}
				<section className="prose mx-auto w-full">
					<AbstractRenderer content={formatAbstract(paper.abstract || "")} />
				</section>
				<div>
					<Subtitle className="mt-0">Resources</Subtitle>
					<div className="flex flex-wrap items-center gap-3">
						<a
							href={`https://huggingface.co/papers/${arxivId}`}
							target="_blank"
							rel="noopener noreferrer"
							className="text-muted-foreground hover:text-foreground cursor-pointer flex items-center gap-2"
						>
							View on Hugging Face
							<ArrowUpRight className="size-4 -ml-1" />
						</a>
						<a
							href={paper.pdfUrl || `https://arxiv.org/pdf/${arxivId}.pdf`}
							target="_blank"
							rel="noopener noreferrer"
							className="text-muted-foreground hover:text-foreground cursor-pointer flex items-center gap-2"
						>
							Read PDF
							<ExternalLink className="size-4 -ml-1" />
						</a>

						<a
							href={`https://arxiv.org/abs/${arxivId}`}
							target="_blank"
							rel="noopener noreferrer"
							className="text-muted-foreground hover:text-foreground cursor-pointer flex items-center gap-2"
						>
							ArXiv
							<ExternalLink className="size-4 -ml-1" />
						</a>
					</div>
				</div>
			</article>
		</main>
	);
}
