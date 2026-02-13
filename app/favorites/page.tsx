"use client";

import { Trash2 } from "lucide-react";
import Link from "next/link";
import { Footer } from "@/components/layout/footer";
import { useFavoritesStore } from "@/lib/stores/favorites-store";
import { cn } from "@/lib/utils";

function formatDate(dateString: string) {
	try {
		return new Date(dateString).toLocaleDateString("en-US", {
			year: "numeric",
			month: "short",
			day: "numeric",
		});
	} catch {
		return "";
	}
}

export default function FavoritesPage() {
	const { getFavorites, removeFavorite, clearAllFavorites } =
		useFavoritesStore();
	const favorites = getFavorites();

	const handleClearAll = () => {
		if (
			confirm(
				"Are you sure you want to clear all favorites? This action cannot be undone.",
			)
		) {
			clearAllFavorites();
		}
	};

	if (favorites.length === 0) {
		return (
			<>
				<main className="mx-auto min-h-screen mb-12 w-full px-6 max-w-5xl">
					<div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
						<h1 className="text-4xl lg:text-5xl font-bold leading-tight tracking-tight mb-4">
							No favorites yet
						</h1>
						<p className="text-muted-foreground text-lg mb-8 max-w-md">
							Start exploring papers and add them to your favorites to see them
							here.
						</p>
						<Link
							href="/"
							className="text-accent hover:text-accent/80 transition-colors"
						>
							Browse papers →
						</Link>
					</div>
				</main>
				<Footer />
			</>
		);
	}

	return (
		<>
			<main className="mx-auto min-h-screen mb-12 w-full px-6 max-w-5xl">
				<div className="flex flex-col gap-8 py-8">
					<div className="flex items-center justify-between border-b border-dashed border-border pb-4">
						<div>
							<h1 className="text-4xl lg:text-5xl font-bold leading-tight tracking-tight">
								Favorites
							</h1>
							<p className="text-muted-foreground mt-2">
								{favorites.length} {favorites.length === 1 ? "paper" : "papers"}
							</p>
							<p className="text-xs text-muted-foreground/70 mt-2">
								Favorites are stored locally on this device only (FOR NOW)
							</p>
						</div>
						<button
							type="button"
							onClick={handleClearAll}
							className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
							aria-label="Clear all favorites"
							title="Clear all favorites"
						>
							<Trash2 className="h-4 w-4" />
							Clear all
						</button>
					</div>

					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-y-16 -mx-6">
						{favorites.map((favorite, index) => (
							<div
								key={favorite.arxivId}
								className={cn(
									"px-6 flex flex-col",
									index % 2 !== 0 && "md:border-l md:border-dashed",
									index % 3 !== 0 && "lg:border-l lg:border-dashed",
									index % 3 === 0 && "lg:border-l-0",
								)}
							>
								<Link
									href={`/p/${favorite.arxivId}`}
									className="flex flex-col overflow-hidden not-prose group gap-5 text-foreground flex-1"
								>
									<div className="flex flex-col gap-2 flex-1">
										<div className="text-xl font-semibold leading-tight line-clamp-2">
											{favorite.title}
										</div>
										<div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground mt-2">
											<span>arXiv:{favorite.arxivId}</span>
											{formatDate(favorite.savedAt) && (
												<>
													<span>·</span>
													<span>Saved {formatDate(favorite.savedAt)}</span>
												</>
											)}
										</div>
									</div>
								</Link>
								<button
									type="button"
									onClick={(e) => {
										e.preventDefault();
										removeFavorite(favorite.arxivId);
									}}
									className="self-start mt-2 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
									aria-label={`Remove ${favorite.title} from favorites`}
									title={`Remove ${favorite.title} from favorites`}
								>
									Remove
								</button>
							</div>
						))}
					</div>
				</div>
			</main>
			<Footer />
		</>
	);
}
