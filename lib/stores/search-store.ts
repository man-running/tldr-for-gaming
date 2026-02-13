import { create } from "zustand";
import type { HuggingFaceSearchResult } from "@/lib/hf-paper-search";

type SearchStage = "idle" | "input" | "results";

interface SearchState {
	searchStage: SearchStage;
	query: string;
	results: HuggingFaceSearchResult[];
	error: string | null;
	totalCount: number;
	isPending: boolean;
}

interface SearchActions {
	setSearchStage: (stage: SearchStage) => void;
	setQuery: (query: string) => void;
	setResults: (results: HuggingFaceSearchResult[]) => void;
	setError: (error: string | null) => void;
	setTotalCount: (count: number) => void;
	setIsPending: (pending: boolean) => void;
	closeSearch: () => void;
	openSearch: () => void;
	reset: () => void;
}

type SearchStore = SearchState & SearchActions;

const initialState: SearchState = {
	searchStage: "idle",
	query: "",
	results: [],
	error: null,
	totalCount: 0,
	isPending: false,
};

export const useSearchStore = create<SearchStore>((set) => ({
	...initialState,
	setSearchStage: (stage) => set({ searchStage: stage }),
	setQuery: (query) => set({ query }),
	setResults: (results) => set({ results }),
	setError: (error) => set({ error }),
	setTotalCount: (totalCount) => set({ totalCount }),
	setIsPending: (isPending) => set({ isPending }),
	closeSearch: () =>
		set({
			searchStage: "idle",
			query: "",
			results: [],
			error: null,
			totalCount: 0,
		}),
	openSearch: () =>
		set((state) => {
			if (state.searchStage === "idle") {
				return { searchStage: "input" };
			}
			return state;
		}),
	reset: () => set(initialState),
}));
