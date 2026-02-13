"use client";

import type { default as DOMPurifyType } from "dompurify";
import type { Route } from "next";
import Link from "next/link";
import React, { useCallback, useEffect, useMemo, useState } from "react";
import { siteConfig } from "@/lib/constants";

interface ContentRendererWithLinksProps {
	html: string;
	isFirstSection?: boolean;
}

export function ContentRendererWithLinks({
	html,
	isFirstSection = false,
}: ContentRendererWithLinksProps) {
	const [cleanHtml, setCleanHtml] = useState<string>("");

	const processPaperLinks = useCallback((content: string): string => {
		let processed = content;
		try {
			const host = new URL(siteConfig.url).host.replace(/\./g, "\\.");
			const domainRegex = new RegExp(`https?:\\/\\/(?:www\\.)?${host}`, "g");
			processed = processed.replace(domainRegex, "");
		} catch (_e) {}
		return processed;
	}, []);

	const reservedMinHeight = useMemo(() => {
		if (cleanHtml) return undefined;
		const approxCharsPerLine = 80;
		const lineHeightPx = isFirstSection ? 30 : 24;
		const lines = Math.ceil(html.length / approxCharsPerLine);
		const px = Math.min(Math.max(lines * lineHeightPx, 80), 600);
		return `${px}px`;
	}, [cleanHtml, html.length, isFirstSection]);

	useEffect(() => {
		let isActive = true;
		(async () => {
			const DOMPurify: typeof DOMPurifyType = (await import("dompurify"))
				.default;
			const htmlWithPaperLinks = processPaperLinks(html);
			const sanitized = DOMPurify.sanitize(htmlWithPaperLinks, {
				ALLOWED_TAGS: [
					"a",
					"b",
					"br",
					"div",
					"em",
					"h1",
					"h2",
					"h3",
					"i",
					"li",
					"ol",
					"p",
					"strong",
					"ul",
				],
				ALLOWED_ATTR: ["href", "target", "rel"],
			});
			if (!isActive) return;
			setCleanHtml(sanitized);
		})();
		return () => {
			isActive = false;
		};
	}, [html, processPaperLinks]);

	const parsedNodes = useMemo(() => {
		if (!cleanHtml) return null;
		const tempDiv = document.createElement("div");
		tempDiv.innerHTML = cleanHtml;

		const parseNode = (node: Node, key: string): React.ReactNode => {
			if (node.nodeType === Node.TEXT_NODE) {
				return node.textContent;
			}
			if (node.nodeType === Node.ELEMENT_NODE) {
				const el = node as Element;
				const tag = el.tagName.toLowerCase();
				const children = Array.from(el.childNodes).map((child, idx) =>
					parseNode(child, `${key}-${idx}`),
				);

				if (tag === "a") {
					const href = el.getAttribute("href") || "";
					if (href.startsWith("/p/")) {
						return (
							<Link
								key={key}
								href={href as Route}
								onClick={() => {
									try {
										sessionStorage.setItem(
											"tldr-scroll",
											String(window.scrollY),
										);
									} catch (_e) {}
								}}
							>
								{children}
							</Link>
						);
					}
					return (
						<a
							key={key}
							href={href}
							target="_blank"
							rel="noopener noreferrer"
							className="text-blue-500 hover:text-blue-600 underline"
						>
							{children}
						</a>
					);
				}

				return React.createElement(tag, { key }, ...children);
			}
			return null;
		};

		return Array.from(tempDiv.childNodes).map((child, idx) =>
			parseNode(child, `n-${idx}`),
		);
	}, [cleanHtml]);

	return (
		<div
			className="prose mx-auto w-full"
			style={
				{
					minHeight: reservedMinHeight,
				} as React.CSSProperties
			}
		>
			{parsedNodes}
		</div>
	);
}
