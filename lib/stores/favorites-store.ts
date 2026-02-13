import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

export interface FavoriteEntry {
	arxivId: string;
	title: string;
	savedAt: string;
}

interface FavoritesStore {
	favorites: FavoriteEntry[];
	addFavorite: (arxivId: string, title: string) => void;
	removeFavorite: (arxivId: string) => void;
	isFavorite: (arxivId: string) => boolean;
	getFavorites: () => FavoriteEntry[];
	clearAllFavorites: () => void;
}

export const useFavoritesStore = create<FavoritesStore>()(
	persist(
		(set, get) => ({
			favorites: [],
			addFavorite: (arxivId: string, title: string) => {
				const { favorites } = get();
				// Check if already favorited
				if (favorites.some((fav) => fav.arxivId === arxivId)) {
					return;
				}
				set({
					favorites: [
						...favorites,
						{
							arxivId,
							title,
							savedAt: new Date().toISOString(),
						},
					],
				});
			},
			removeFavorite: (arxivId: string) => {
				const { favorites } = get();
				set({
					favorites: favorites.filter((fav) => fav.arxivId !== arxivId),
				});
			},
			isFavorite: (arxivId: string) => {
				const { favorites } = get();
				return favorites.some((fav) => fav.arxivId === arxivId);
			},
			getFavorites: () => {
				const { favorites } = get();
				return [...favorites].sort(
					(a, b) =>
						new Date(b.savedAt).getTime() - new Date(a.savedAt).getTime(),
				);
			},
			clearAllFavorites: () => {
				set({ favorites: [] });
			},
		}),
		{
			name: "tldr-favorites",
			storage: createJSONStorage(() => localStorage),
		},
	),
);
