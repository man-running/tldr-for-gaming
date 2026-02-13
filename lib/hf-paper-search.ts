const SEARCH_PROXY_URL = "/hf/papers";

interface HuggingFaceResponseItem {
	paper?: {
		id: string;
		title: string;
		summary?: string;
		publishedAt?: string;
	};
}

export interface HuggingFaceSearchResult {
	id: string;
	title: string;
	summary: string;
	publishedAt: string;
}

interface QueryEmbeddingResponse {
	queryEmbedding: number[];
}

function isQueryEmbeddingResponse(
	data: unknown,
): data is QueryEmbeddingResponse {
	return (
		typeof data === "object" &&
		data !== null &&
		"queryEmbedding" in data &&
		Array.isArray((data as QueryEmbeddingResponse).queryEmbedding)
	);
}

function isHuggingFaceResponseItem(
	item: unknown,
): item is HuggingFaceResponseItem {
	return (
		typeof item === "object" &&
		item !== null &&
		"paper" in item &&
		typeof (item as HuggingFaceResponseItem).paper === "object" &&
		(item as HuggingFaceResponseItem).paper !== null
	);
}

async function fetchFromHuggingFace(
	query: string,
	signal?: AbortSignal,
): Promise<HuggingFaceSearchResult[]> {
	const url = `${SEARCH_PROXY_URL}?q=${encodeURIComponent(query)}`;

	try {
		const response = await fetch(url, { signal });
		if (!response.ok) {
			throw new Error(`Search failed: ${response.status}`);
		}
		const payload: unknown = await response.json();
		if (!Array.isArray(payload)) return [];

		return payload
			.filter(isHuggingFaceResponseItem)
			.filter((item) => item.paper?.id && item.paper?.title)
			.map((item) => ({
				id: item.paper?.id ?? "",
				title: item.paper?.title ?? "",
				summary: item.paper?.summary || "",
				publishedAt: item.paper?.publishedAt || "",
			}));
	} catch (err) {
		if (err instanceof Error && err.name === "AbortError") throw err;
		throw new Error(
			err instanceof Error ? err.message : "Failed to fetch search results",
		);
	}
}

export async function generateQueryEmbedding(
	query: string,
	signal?: AbortSignal,
): Promise<Float32Array | null> {
	try {
		const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`, {
			signal,
		});
		if (!response.ok) return null;
		const data: unknown = await response.json();
		if (!isQueryEmbeddingResponse(data) || !data.queryEmbedding.length) {
			return null;
		}
		return new Float32Array(data.queryEmbedding);
	} catch {
		return null;
	}
}

function isHuggingFaceSearchResult(
	item: unknown,
): item is HuggingFaceSearchResult {
	return (
		typeof item === "object" &&
		item !== null &&
		"id" in item &&
		"title" in item &&
		typeof (item as HuggingFaceSearchResult).id === "string" &&
		typeof (item as HuggingFaceSearchResult).title === "string"
	);
}

function getYearFromDate(publishedAt?: string): string {
	if (!publishedAt) {
		return "Older";
	}
	const date = new Date(publishedAt);
	if (Number.isNaN(date.getTime())) {
		return "Older";
	}
	return date.getFullYear().toString();
}

function _groupByYear(
	results: HuggingFaceSearchResult[],
): Record<string, HuggingFaceSearchResult[]> {
	return results.reduce<Record<string, HuggingFaceSearchResult[]>>(
		(acc, paper) => {
			const year = getYearFromDate(paper.publishedAt);
			if (!acc[year]) {
				acc[year] = [];
			}
			acc[year].push(paper);
			return acc;
		},
		{},
	);
}

export async function searchPapers(
	query: string,
	signal?: AbortSignal,
	onInitialResults?: (
		firstResult: HuggingFaceSearchResult | null,
		totalCount: number,
	) => void,
	queryEmbedding?: Float32Array,
): Promise<HuggingFaceSearchResult[]> {
	const trimmedQuery = query.trim();
	if (!trimmedQuery) {
		return [];
	}

	// Start query embedding generation in parallel with HuggingFace fetch
	// If not provided, generate it now (don't wait for HuggingFace to finish)
	const embeddingPromise = queryEmbedding
		? Promise.resolve(queryEmbedding)
		: generateQueryEmbedding(trimmedQuery, signal);

	// Fetch HuggingFace results and generate query embedding in parallel
	const [results, generatedEmbedding] = await Promise.all([
		fetchFromHuggingFace(trimmedQuery, signal),
		embeddingPromise,
	]);

	// Use provided embedding or the one we just generated
	const finalQueryEmbedding = queryEmbedding || generatedEmbedding;

	// Show only first result + total count immediately (not all results)
	if (onInitialResults) {
		const firstResult = results.length > 0 ? results[0] : null;
		onInitialResults(firstResult, results.length);
	}

	// Separate first result (hero) from rest
	const heroPaper = results[0];
	const remaining = results.slice(1);

	if (remaining.length === 0) {
		return results;
	}

	// Security: Limit results array size before sending
	const MAX_RESULTS = 1000;
	const limitedRemaining = remaining.slice(0, MAX_RESULTS);

	// Send POST request immediately after receiving HuggingFace data
	const embeddingArray = finalQueryEmbedding
		? Array.from(finalQueryEmbedding)
		: undefined;

	try {
		const response = await fetch("/api/search", {
			method: "POST",
			signal,
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({
				query: trimmedQuery,
				results: limitedRemaining,
				queryEmbedding: embeddingArray,
			}),
		});

		// Fallback on error status codes (400-599)
		if (!response.ok) {
			return results;
		}

		const reranked: unknown = await response.json();
		if (!Array.isArray(reranked)) {
			return results;
		}

		const validated = reranked.filter(isHuggingFaceSearchResult);
		const rerankedRemaining = validated.length > 0 ? validated : remaining;

		const finalResults: HuggingFaceSearchResult[] = heroPaper
			? [heroPaper]
			: [];
		finalResults.push(...rerankedRemaining);
		return finalResults;
	} catch {
		// Network errors or JSON parsing failures - fallback to original results
		return results;
	}
}
