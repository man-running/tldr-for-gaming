"use client";

import { ArrowRightIcon, Calendar, Layers3, X } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { memo, useCallback, useMemo, useState } from "react";
import type { HuggingFaceSearchResult } from "@/lib/hf-paper-search";
import { useSearchStore } from "@/lib/stores/search-store";
import { cn } from "@/lib/utils";

interface PaperSearchPanelProps {
	onNavigate?: () => void;
}

function PaperSearchPanel({ onNavigate }: PaperSearchPanelProps) {
	const {
		searchStage,
		query,
		results,
		error,
		totalCount,
		isPending,
		closeSearch,
	} = useSearchStore();

	const trimmedQuery = query.trim();
	const hasQuery = query.length > 0;
	const isVisible = searchStage !== "idle" && hasQuery;
	const loading = isPending;
	const [expandedYears, setExpandedYears] = useState<Set<string>>(new Set());
	const maxPapersPerYear = 12;

	const getYearFromDate = useCallback((publishedAt?: string) => {
		if (!publishedAt) return "Older";
		const year = new Date(publishedAt).getFullYear();
		return Number.isNaN(year) ? "Older" : year.toString();
	}, []);

	const heroPaper = results[0];
	const remaining = results.slice(1);

	const groupedByYear = useMemo(
		() =>
			remaining.reduce<Record<string, HuggingFaceSearchResult[]>>(
				(acc, paper) => {
					const year = getYearFromDate(paper.publishedAt);
					if (!acc[year]) acc[year] = [];
					acc[year].push(paper);
					return acc;
				},
				{},
			),
		[remaining, getYearFromDate],
	);

	const sortedYearBuckets = useMemo(
		() =>
			Object.entries(groupedByYear).sort(([a], [b]) => {
				if (a === "Older") return 1;
				if (b === "Older") return -1;
				return Number(b) - Number(a);
			}),
		[groupedByYear],
	);

	const yearNavItems = useMemo(
		() =>
			sortedYearBuckets.map(([year, papers]) => ({
				year,
				count: papers.length,
			})),
		[sortedYearBuckets],
	);

	if (!isVisible) {
		return null;
	}

	const showEmptyState =
		!loading && !error && trimmedQuery.length > 0 && results.length === 0;

	return (
		<div className="fixed inset-x-0 bottom-0 top-[75px] sm:top-[93px] z-30 bg-background/95 backdrop-blur-sm flex flex-col overflow-hidden">
			<div className="flex-1 overflow-y-auto">
				<div className="mx-auto w-full max-w-5xl px-6 py-12 lg:py-20 flex flex-col gap-8">
					<div className="flex items-start justify-between">
						<div>
							<h1 className="text-4xl lg:text-5xl font-bold leading-tight tracking-tight">
								"{trimmedQuery}"
							</h1>
							{trimmedQuery.length > 0 && (
								<div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground mt-4">
									<span className="inline-flex items-center gap-1">
										<Layers3 className="h-4 w-4" />
										{totalCount !== undefined && totalCount > 0
											? `${totalCount} paper${totalCount === 1 ? "" : "s"}`
											: results.length > 0
												? `${results.length} paper${results.length === 1 ? "" : "s"}`
												: "Waiting for matches"}
									</span>
									<span className="inline-flex items-center gap-1 text-xs text-muted-foreground/70">
										Search powered by takara <span className="italic">DS1</span>
										<Image
											src="/icon.svg"
											alt=""
											width={16}
											height={12}
											className="object-contain opacity-60"
										/>
										— the world's fastest embedding service
									</span>
								</div>
							)}
						</div>
						<button
							type="button"
							onClick={closeSearch}
							className="text-muted-foreground hover:text-foreground transition-colors"
							aria-label="Close search"
						>
							<X className="h-5 w-5" />
						</button>
					</div>

					{yearNavItems.length > 0 && (
						<div className="flex flex-wrap gap-2 overflow-x-auto py-4 border-y border-dashed border-border">
							{yearNavItems.map((item) => (
								<button
									key={item.year}
									type="button"
									onClick={() => {
										const el = document.getElementById(`year-${item.year}`);
										el?.scrollIntoView({ behavior: "smooth", block: "start" });
									}}
									className="text-sm text-muted-foreground hover:text-foreground whitespace-nowrap transition-colors"
								>
									{item.year} · {item.count}
								</button>
							))}
						</div>
					)}

					<div className="space-y-8 pt-6">
						{trimmedQuery.length === 0 && !loading && (
							<div className="flex items-center justify-center h-full">
								<p className="text-base text-muted-foreground text-center max-w-lg">
									Start typing in the navbar search to discover papers
									instantly.
								</p>
							</div>
						)}

						{error && (
							<div className="flex items-center justify-center h-full">
								<p className="text-base text-destructive text-center">
									{error}
								</p>
							</div>
						)}

						{showEmptyState && (
							<div className="flex items-center justify-center h-full">
								<p className="text-base text-muted-foreground text-center">
									No papers found for this query yet.
								</p>
							</div>
						)}

						{heroPaper && (
							<Link
								href={`/p/${heroPaper.id}`}
								className="flex flex-col overflow-hidden not-prose group gap-5 text-foreground"
								onClick={() => {
									onNavigate?.();
									closeSearch();
								}}
							>
								<div className="flex flex-col gap-2 flex-1">
									<p className="text-sm text-muted-foreground inline-flex items-center gap-2">
										Fastest match{" "}
										<ArrowRightIcon className="size-4 shrink-0 opacity-0 group-hover:opacity-100 -translate-x-1 group-hover:translate-x-0 transition-all ml-auto" />
									</p>
									<div className="text-xl font-semibold leading-tight line-clamp-2">
										{heroPaper.title}
									</div>
									{heroPaper.summary && (
										<p className="text-sm text-muted-foreground line-clamp-3">
											{heroPaper.summary}
										</p>
									)}
									<div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground mt-2">
										{heroPaper.publishedAt && (
											<span className="inline-flex items-center gap-1">
												<Calendar className="h-4 w-4" />
												{new Date(heroPaper.publishedAt).toLocaleDateString(
													"en-US",
													{
														year: "numeric",
														month: "short",
														day: "numeric",
													},
												)}
											</span>
										)}
										{heroPaper.publishedAt && <span>·</span>}
										<span>arXiv:{heroPaper.id}</span>
									</div>
								</div>
							</Link>
						)}

						{sortedYearBuckets.length > 0 && (
							<div className="space-y-8">
								{sortedYearBuckets.map(([year, papers]) => {
									const isExpanded = expandedYears.has(year);
									const hasMorePapers = papers.length > maxPapersPerYear;
									const visiblePapers = isExpanded
										? papers
										: papers.slice(0, maxPapersPerYear);
									const remainingCount = hasMorePapers
										? papers.length - maxPapersPerYear
										: 0;

									const toggleExpanded = () => {
										setExpandedYears((prev) => {
											const next = new Set(prev);
											if (next.has(year)) {
												next.delete(year);
											} else {
												next.add(year);
											}
											return next;
										});
									};

									return (
										<section
											key={year}
											id={`year-${year}`}
											className="space-y-6 scroll-mt-28"
										>
											<div className="flex items-center justify-between border-b border-dashed border-border pb-2">
												<h2 className="text-base text-muted-foreground font-medium">
													{year}
												</h2>
												<span className="text-sm text-muted-foreground">
													{papers.length} paper{papers.length === 1 ? "" : "s"}
												</span>
											</div>
											<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-y-16 -mx-6">
												{visiblePapers.map((paper, index) => (
													<div
														key={paper.id}
														className={cn(
															"px-6",
															index % 2 !== 0 && "md:border-l md:border-dashed",
															index % 3 !== 0 && "lg:border-l lg:border-dashed",
															index % 3 === 0 && "lg:border-l-0",
														)}
													>
														<Link
															href={`/p/${paper.id}`}
															className="flex flex-col overflow-hidden not-prose group gap-5 text-foreground"
															onClick={() => {
																onNavigate?.();
																closeSearch();
															}}
														>
															<div className="flex flex-col gap-2 flex-1">
																<div className="text-xl font-semibold leading-tight line-clamp-2">
																	{paper.title}
																</div>
																{paper.summary && (
																	<p className="text-sm text-muted-foreground line-clamp-3">
																		{paper.summary}
																	</p>
																)}
																<div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground mt-2">
																	{paper.publishedAt && (
																		<span className="inline-flex items-center gap-1">
																			<Calendar className="h-4 w-4" />
																			{new Date(
																				paper.publishedAt,
																			).toLocaleDateString("en-US", {
																				year: "numeric",
																				month: "short",
																				day: "numeric",
																			})}
																		</span>
																	)}
																	{paper.publishedAt && <span>·</span>}
																	<span>arXiv:{paper.id}</span>
																</div>
															</div>
														</Link>
													</div>
												))}
											</div>
											{hasMorePapers && (
												<div className="flex justify-center pt-2">
													<button
														type="button"
														onClick={toggleExpanded}
														className="text-muted-foreground hover:text-foreground cursor-pointer text-sm"
														aria-label={
															isExpanded
																? "Show fewer papers"
																: `Show ${remainingCount} more papers`
														}
													>
														{isExpanded
															? "show less"
															: `and ${remainingCount} more`}
													</button>
												</div>
											)}
										</section>
									);
								})}
							</div>
						)}

						{loading && (
							<div className="space-y-md">
								{["alpha", "beta", "gamma", "delta"].map((id) => (
									<div key={id} className="space-y-sm">
										<div className="skeleton-heading w-2/3" />
										<div className="skeleton-text w-full" />
										<div className="skeleton-text w-3/4" />
									</div>
								))}
							</div>
						)}
					</div>
				</div>
			</div>
		</div>
	);
}

export const PaperSearchPanelMemo = memo(PaperSearchPanel);
