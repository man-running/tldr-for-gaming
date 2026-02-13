"use client";

import { Heart } from "lucide-react";
import { useFavoritesStore } from "@/lib/stores/favorites-store";
import { usePaper } from "./paper-data-provider";

interface FavoriteButtonProps {
	arxivId: string;
}

export function FavoriteButton({ arxivId }: FavoriteButtonProps) {
	const { paper } = usePaper();
	const { isFavorite, addFavorite, removeFavorite } = useFavoritesStore();
	const favorited = isFavorite(arxivId);

	const handleToggle = () => {
		if (favorited) {
			removeFavorite(arxivId);
		} else {
			if (paper?.title) {
				addFavorite(arxivId, paper.title);
			}
		}
	};

	return (
		<button
			type="button"
			onClick={handleToggle}
			className={`cursor-pointer transition-colors ${favorited ? "text-accent" : "text-muted-foreground hover:text-foreground"}`}
			aria-label={favorited ? "Remove from favorites" : "Add to favorites"}
			title={favorited ? "Remove from favorites" : "Add to favorites"}
		>
			<Heart
				className={`h-5 w-5 ${favorited ? "fill-accent text-accent" : ""}`}
			/>
		</button>
	);
}
