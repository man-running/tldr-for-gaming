"use client";

import { useEffect } from "react";
import posthog from "posthog-js";
import { PostHogProvider as PHProvider } from "posthog-js/react";

export function PostHogProvider({ children }: { children: React.ReactNode }) {
	useEffect(() => {
		const posthogKey = process.env.NEXT_PUBLIC_POSTHOG_KEY;
		if (posthogKey && typeof window !== "undefined") {
			posthog.init(posthogKey, {
				api_host: "/ingest",
				ui_host: "https://eu.posthog.com",
				loaded: (posthog) => {
					if (process.env.NODE_ENV === "development") {
						posthog.debug();
					}
				},
				capture_pageview: true,
				capture_pageleave: true,
			});
		}
	}, []);

	return <PHProvider client={posthog}>{children}</PHProvider>;
}
