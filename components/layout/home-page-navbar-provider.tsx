"use client";

import { createContext, type ReactNode } from "react";
import { useFeed } from "@/components/feed/feed-data-provider";

interface NavbarFeedData {
	currentDate?: string;
	navigateToDate?: (date: string) => void;
	availableDates?: string[];
}

export const NavbarFeedContext = createContext<NavbarFeedData | undefined>(
	undefined,
);

export function HomePageNavbarProvider({ children }: { children: ReactNode }) {
	const { currentDate, navigateToDate, availableDates } = useFeed();

	const value: NavbarFeedData = {
		currentDate: currentDate ?? undefined,
		navigateToDate,
		availableDates,
	};

	return (
		<NavbarFeedContext.Provider value={value}>
			{children}
		</NavbarFeedContext.Provider>
	);
}
