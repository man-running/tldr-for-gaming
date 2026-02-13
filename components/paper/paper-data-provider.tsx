"use client";

import {
	createContext,
	type ReactNode,
	useCallback,
	useContext,
	useState,
} from "react";
import type { PaperData } from "@/types/api";

interface FeaturedImage {
	url: string;
	alt: string;
}

interface PaperContextType {
	paper: PaperData | null;
	loading: boolean;
	error: string | null;
	featuredImage: FeaturedImage | null;
	refetch: () => void;
}

const PaperContext = createContext<PaperContextType | undefined>(undefined);

interface PaperDataProviderProps {
	children: ReactNode;
	initialPaper?: PaperData;
	initialFeaturedImage?: FeaturedImage | null;
}

export function PaperDataProvider({
	children,
	initialPaper,
	initialFeaturedImage,
}: PaperDataProviderProps) {
	const [paper] = useState<PaperData | null>(initialPaper || null);
	const [loading] = useState(false); // No loading since data is pre-fetched
	const [error] = useState<string | null>(
		initialPaper ? null : "Paper not found",
	);
	const [featuredImage] = useState<FeaturedImage | null>(
		initialFeaturedImage ?? null,
	);

	// Optional refetch function for future use (though not needed with server-side data)
	const fetchPaper = useCallback(async () => {
		// No-op since data is server-side fetched and cached
	}, []);

	return (
		<PaperContext.Provider
			value={{ paper, loading, error, featuredImage, refetch: fetchPaper }}
		>
			{children}
		</PaperContext.Provider>
	);
}

export function usePaper() {
	const context = useContext(PaperContext);
	if (!context)
		throw new Error("usePaper must be used within PaperDataProvider");
	return context;
}
