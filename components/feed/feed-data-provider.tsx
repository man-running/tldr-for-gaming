"use client";

import {
	createContext,
	type ReactNode,
	useCallback,
	useContext,
	useEffect,
	useState,
} from "react";
import { useFeedStore } from "@/lib/stores/feed-store";
import type { RssFeed } from "@/types/feed";

interface FeedContextType {
	feed: RssFeed | null;
	loading: boolean;
	error: string | null;
	currentDate: string | null;
	navigateToDate: (date: string) => void;
	availableDates: string[];
}

const FeedContext = createContext<FeedContextType | undefined>(undefined);

// Helper function to fetch blob URLs with timeout (plain fetch, no options to avoid CORS preflight)
const fetchBlobWithTimeout = async (
	url: string,
	timeoutMs = 10000,
): Promise<Response> => {
	const controller = new AbortController();
	const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

	try {
		// Plain fetch with no options - Vercel Blob doesn't support CORS preflight
		const response = await fetch(url, {
			signal: controller.signal,
		});
		clearTimeout(timeoutId);
		return response;
	} catch (err) {
		clearTimeout(timeoutId);
		if (err instanceof Error && err.name === "AbortError") {
			throw new Error("Request timeout - please try again");
		}
		throw err;
	}
};

// Helper function to fetch API endpoints with timeout and options
const fetchWithTimeout = async (
	url: string,
	options: RequestInit = {},
	timeoutMs = 10000,
): Promise<Response> => {
	const controller = new AbortController();
	const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

	try {
		const response = await fetch(url, {
			...options,
			signal: controller.signal,
		});
		clearTimeout(timeoutId);
		return response;
	} catch (err) {
		clearTimeout(timeoutId);
		if (err instanceof Error && err.name === "AbortError") {
			throw new Error("Request timeout - please try again");
		}
		throw err;
	}
};

export function FeedDataProvider({ children }: { children: ReactNode }) {
	const [feed, setFeed] = useState<RssFeed | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [availableDates, setAvailableDates] = useState<string[]>([]);
	const { selectedDate, setSelectedDate } = useFeedStore();
	const currentDate = selectedDate;

	// Initialize selectedDate with latest if not set
	useEffect(() => {
		if (selectedDate) return;

		let isActive = true;
		(async () => {
			try {
				// Add cache-busting to prevent Safari caching issues
				const cacheBuster = `?t=${Date.now()}`;
				const datesIndexURL = `https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com/tldr-summaries/dates-index.json${cacheBuster}`;
				const datesResponse = await fetchBlobWithTimeout(datesIndexURL, 8000);
				if (datesResponse.ok && isActive) {
					const datesIndex = await datesResponse.json();
					const latestDate = datesIndex.dates?.[0];
					if (latestDate) {
						setSelectedDate(latestDate);
					}
				}
			} catch (_err) {
				// Silently fail
			}
		})();
		return () => {
			isActive = false;
		};
	}, [selectedDate, setSelectedDate]);

	// Fetch available dates for calendar (lazy, only when needed)
	// Fetch directly from blob storage: tldr-summaries/dates-index.json
	useEffect(() => {
		let isActive = true;
		(async () => {
			try {
				// Add cache-busting query parameter to prevent Safari caching issues
				const cacheBuster = `?t=${Date.now()}`;
				const blobURL = `https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com/tldr-summaries/dates-index.json${cacheBuster}`;
				const res = await fetchBlobWithTimeout(blobURL, 8000);
				if (!res.ok) {
					// Fallback to API endpoint if blob fetch fails
					const base = process.env.NEXT_PUBLIC_BASE_URL || "";
					const apiRes = await fetchWithTimeout(
						`${base}/api/archive${cacheBuster}`,
						{ cache: "no-store" },
						8000,
					);
					if (!apiRes.ok) throw new Error("Failed to fetch archive dates");
					const json = await apiRes.json();
					if (isActive && json.dates && Array.isArray(json.dates)) {
						setAvailableDates(json.dates);
					}
					return;
				}
				const indexFile = await res.json();
				if (isActive && indexFile.dates && Array.isArray(indexFile.dates)) {
					setAvailableDates(indexFile.dates);
				} else if (isActive) {
					// If response doesn't have expected structure, try fallback
					console.warn("Unexpected dates index structure, trying API fallback");
					const base = process.env.NEXT_PUBLIC_BASE_URL || "";
					const apiRes = await fetchWithTimeout(
						`${base}/api/archive${cacheBuster}`,
						{ cache: "no-store" },
						8000,
					);
					if (apiRes.ok) {
						const json = await apiRes.json();
						if (isActive && json.dates && Array.isArray(json.dates)) {
							setAvailableDates(json.dates);
						}
					}
				}
			} catch (err) {
				// Log error for debugging but don't break the app
				console.error("Failed to fetch available dates:", err);
				if (isActive) {
					setAvailableDates([]);
				}
			}
		})();
		return () => {
			isActive = false;
		};
	}, []);

	const fetchFeed = useCallback(async (date: string | null = null) => {
		try {
			setLoading(true);
			setError(null);

			// Validate date format if provided
			if (date && !/^\d{4}-\d{2}-\d{2}$/.test(date)) {
				setError("Invalid date format");
				setLoading(false);
				return;
			}

			const base = process.env.NEXT_PUBLIC_BASE_URL || "";

			// For archive dates, fetch directly from blob storage
			// Blob URL pattern: https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com/tldr-feeds/{date}.json
			if (date) {
				const blobURL = `https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com/tldr-feeds/${date}.json`;
				try {
					const response = await fetchBlobWithTimeout(blobURL, 10000);
					if (!response.ok) {
						if (response.status === 404) {
							setFeed(null);
							setError(`No data available for ${date}`);
							setLoading(false);
							return;
						}
						throw new Error(`Blob fetch failed with status ${response.status}`);
					}
					const data = (await response.json()) as RssFeed;
					if (!data?.items) {
						setError("Feed data is invalid or empty");
						setFeed(null);
						setLoading(false);
						return;
					}
					setFeed(data);
					setLoading(false);
					return;
				} catch (err) {
					// Fallback to API endpoint if blob fetch fails or times out
					console.error(
						"Failed to fetch from blob URL, falling back to API:",
						err,
					);
					try {
						const response = await fetchWithTimeout(
							`${base}/api/archive?date=${encodeURIComponent(date)}`,
							{ cache: "no-store" },
							10000,
						);
						if (!response.ok) {
							if (response.status === 404) {
								setFeed(null);
								setError(`No data available for ${date}`);
								setLoading(false);
								return;
							}
							throw new Error(
								`Feed fetch failed with status ${response.status}`,
							);
						}
						const data = (await response.json()) as RssFeed;
						if (!data?.items) {
							setError("Feed data is invalid or empty");
							setFeed(null);
							setLoading(false);
							return;
						}
						setFeed(data);
						setLoading(false);
						return;
					} catch (fallbackErr) {
						// If fallback also fails, throw to outer catch
						throw fallbackErr;
					}
				}
			}

			// For current feed (no date specified), fetch latest from blob storage
			// Same logic as initial mount - get latest date and fetch feed
			const cacheBuster = `?t=${Date.now()}`;
			const datesIndexURL = `https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com/tldr-summaries/dates-index.json${cacheBuster}`;
			try {
				const datesResponse = await fetchBlobWithTimeout(datesIndexURL, 8000);
				if (datesResponse.ok) {
					const datesIndex = await datesResponse.json();
					const latestDate = datesIndex.dates?.[0];

					if (latestDate) {
						const feedBlobURL = `https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com/tldr-feeds/${latestDate}.json`;
						const feedResponse = await fetchBlobWithTimeout(feedBlobURL, 10000);

						if (feedResponse.ok) {
							const data = (await feedResponse.json()) as RssFeed;
							if (!data?.items) {
								setError("Feed data is invalid or empty");
								setFeed(null);
								setLoading(false);
								return;
							}
							setFeed(data);
							setLoading(false);
							return;
						}
					}
				}
			} catch (err) {
				console.error(
					"Failed to fetch from blob storage, falling back to API:",
					err,
				);
			}

			// Fallback to API endpoint
			const response = await fetchWithTimeout(
				`${base}/api/feed`,
				{ cache: "no-store" },
				10000,
			);
			if (!response.ok) {
				throw new Error(`Feed fetch failed with status ${response.status}`);
			}

			const data = (await response.json()) as RssFeed;
			if (!data?.items) {
				setError("Feed data is invalid or empty");
				setFeed(null);
				setLoading(false);
				return;
			}

			setFeed(data);
			setLoading(false);
		} catch (err) {
			const errorMsg =
				err instanceof Error ? err.message : "Unable to load feed";
			setError(errorMsg);
			setFeed(null);
		} finally {
			setLoading(false);
		}
	}, []);

	// Fetch feed when selectedDate changes, or load latest if no date is selected
	useEffect(() => {
		if (selectedDate) {
			void fetchFeed(selectedDate);
		} else {
			// If no date is selected, load the latest feed
			// This ensures the feed loads even if sessionStorage fails
			void fetchFeed(null);
		}
	}, [selectedDate, fetchFeed]);

	const navigateToDate = useCallback(
		(date: string) => {
			setSelectedDate(date);
		},
		[setSelectedDate],
	);

	return (
		<FeedContext.Provider
			value={{
				feed,
				loading,
				error,
				currentDate,
				navigateToDate,
				availableDates,
			}}
		>
			{children}
		</FeedContext.Provider>
	);
}

export function useFeed() {
	const context = useContext(FeedContext);
	if (!context) throw new Error("useFeed must be used within FeedDataProvider");
	return context;
}
