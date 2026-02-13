import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

interface FeedStore {
	selectedDate: string | null;
	setSelectedDate: (date: string | null) => void;
}

// Safe sessionStorage wrapper that handles Safari mobile storage failures
const safeSessionStorage = {
	getItem: (name: string): string | null => {
		try {
			return sessionStorage.getItem(name);
		} catch {
			return null;
		}
	},
	setItem: (name: string, value: string): void => {
		try {
			sessionStorage.setItem(name, value);
		} catch {
			// Silently fail - storage unavailable (e.g., Safari private mode)
		}
	},
	removeItem: (name: string): void => {
		try {
			sessionStorage.removeItem(name);
		} catch {
			// Silently fail
		}
	},
};

export const useFeedStore = create<FeedStore>()(
	persist(
		(set) => ({
			selectedDate: null,
			setSelectedDate: (date) => set({ selectedDate: date }),
		}),
		{
			name: "tldr-feed-date",
			storage: createJSONStorage(() => safeSessionStorage),
		},
	),
);
