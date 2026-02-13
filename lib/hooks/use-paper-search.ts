import { useCallback, useRef, useTransition } from "react";
import type { HuggingFaceSearchResult } from "@/lib/hf-paper-search";
import { generateQueryEmbedding, searchPapers } from "@/lib/hf-paper-search";
import { useSearchStore } from "@/lib/stores/search-store";

interface UsePaperSearchOptions {
	debouncedQuery: string;
	searchStage: "idle" | "input" | "results";
	hasQuery: boolean;
}

export function usePaperSearch({
	debouncedQuery,
	searchStage,
	hasQuery,
}: UsePaperSearchOptions) {
	const [isPending, startTransition] = useTransition();
	const abortControllerRef = useRef<AbortController | null>(null);
	const queryEmbeddingRef = useRef<Float32Array | null>(null);
	const lastSearchedQueryRef = useRef<string>("");

	const { setResults, setError, setTotalCount, setSearchStage, setIsPending } =
		useSearchStore();

	const performSearch = useCallback(() => {
		if (!debouncedQuery || searchStage === "idle") {
			if (!hasQuery) {
				setResults([]);
				if (searchStage === "results") setSearchStage("input");
			}
			lastSearchedQueryRef.current = "";
			return;
		}

		if (lastSearchedQueryRef.current === debouncedQuery) return;
		lastSearchedQueryRef.current = debouncedQuery;

		abortControllerRef.current?.abort();
		const controller = new AbortController();
		abortControllerRef.current = controller;

		startTransition(async () => {
			setIsPending(true);
			try {
				const response = await searchPapers(
					debouncedQuery,
					controller.signal,
					(firstResult, totalCount) => {
						if (!controller.signal.aborted) {
							setTotalCount(totalCount);
							setError(null);
							setSearchStage("results");
							setResults(firstResult ? [firstResult] : []);
						}
					},
					queryEmbeddingRef.current || undefined,
				);

				if (!controller.signal.aborted) {
					setResults(response);
					setError(null);
					setSearchStage("results");
				}
			} catch (err) {
				if (
					controller.signal.aborted ||
					(err instanceof Error && err.name === "AbortError")
				) {
					return;
				}
				const errorMessage =
					err instanceof Error && err.message.includes("Network error")
						? "Network error. Please check your connection."
						: err instanceof Error && err.message.includes("Search failed")
							? "Search service unavailable. Please try again."
							: "Something went wrong. Please try again.";
				setError(errorMessage);
				setResults([]);
			} finally {
				setIsPending(false);
			}
		});
	}, [
		debouncedQuery,
		searchStage,
		hasQuery,
		setResults,
		setError,
		setTotalCount,
		setSearchStage,
		setIsPending,
	]);

	const loadQueryEmbedding = useCallback(() => {
		if (!debouncedQuery || searchStage === "idle") {
			queryEmbeddingRef.current = null;
			return;
		}

		const controller = new AbortController();
		generateQueryEmbedding(debouncedQuery, controller.signal).then(
			(embedding) => {
				if (!controller.signal.aborted && embedding) {
					queryEmbeddingRef.current = embedding;
				}
			},
		);
		return () => controller.abort();
	}, [debouncedQuery, searchStage]);

	const abort = useCallback(() => {
		abortControllerRef.current?.abort();
	}, []);

	return {
		isPending,
		performSearch,
		loadQueryEmbedding,
		abort,
	};
}
