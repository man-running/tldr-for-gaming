"use client";

import { useEffect } from "react";

export function RestoreScroll() {
	useEffect(() => {
		try {
			const raw = sessionStorage.getItem("tldr-scroll");
			if (!raw) return;
			const y = Number(raw);
			requestAnimationFrame(() => window.scrollTo(0, y));
			sessionStorage.removeItem("tldr-scroll");
		} catch {
			// Silently fail if sessionStorage is unavailable (e.g., Safari private mode)
		}
	}, []);
	return null;
}
