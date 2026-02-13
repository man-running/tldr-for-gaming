"use client";

import { Heart } from "lucide-react";
import Link from "next/link";
import { useFavoritesStore } from "@/lib/stores/favorites-store";

export function FavoritesMenu() {
	const hasFavorites = useFavoritesStore((state) => state.favorites.length > 0);

	if (!hasFavorites) {
		return null;
	}

	return (
		<Link
			href="/favorites"
			className="icon-btn cursor-pointer"
			aria-label="View favorites"
			title="View favorites"
		>
			<Heart className="h-5 w-5 icon-color" />
		</Link>
	);
}
