"use client";

import { useState } from "react";

interface AuthorDisplayProps {
	authors: string[];
}

export function AuthorDisplay({ authors }: AuthorDisplayProps) {
	const [isExpanded, setIsExpanded] = useState(false);

	const maxVisibleAuthors = 5;
	const hasMoreAuthors = authors.length > maxVisibleAuthors;
	const remainingCount = hasMoreAuthors
		? authors.length - maxVisibleAuthors
		: 0;

	const collapsedAuthorsText = hasMoreAuthors
		? authors.slice(0, maxVisibleAuthors).join(", ")
		: authors.join(", ");
	const fullAuthorsText = authors.join(", ");

	const containerClass = "font-lato text-foreground";
	const actionClass =
		"text-accent hover:text-accent/80 hover:underline cursor-pointer";

	if (!hasMoreAuthors) {
		return <div className={containerClass}>{fullAuthorsText}</div>;
	}

	return (
		<div className={containerClass}>
			{isExpanded ? (
				<>
					<span>{fullAuthorsText}</span>
					<span> </span>
					<button
						type="button"
						onClick={() => setIsExpanded(false)}
						className={actionClass}
						aria-label="Show fewer authors"
					>
						(show less)
					</button>
				</>
			) : (
				<>
					<span>{collapsedAuthorsText}</span>
					<span>{", "}</span>
					<button
						type="button"
						onClick={() => setIsExpanded(true)}
						className={actionClass}
						aria-label={`Show all ${authors.length} authors`}
					>
						and {remainingCount} more
					</button>
				</>
			)}
		</div>
	);
}
