"use client";

import { useFeed } from "./feed-data-provider";
import { FeedError } from "./feed-error";
import { FeedItem } from "./feed-item-simplified";

function FeedSkeleton() {
	return (
		<div className="space-y-md">
			{[1, 2, 3].map((i) => (
				<div key={i} className="rounded-xl p-6 shadow-sm space-y-md">
					<div className="skeleton-heading w-3/4" />
					<div className="space-y-sm">
						<div className="skeleton-text" />
						<div className="skeleton-text w-5/6" />
					</div>
				</div>
			))}
		</div>
	);
}

export function FeedContent() {
	const { feed, loading, error } = useFeed();

	if (loading) {
		return <FeedSkeleton />;
	}

	if (error) {
		return <FeedError />;
	}

	if (!feed) {
		return <FeedError />;
	}

	if (!feed.items || feed.items.length === 0) {
		return (
			<div className="rounded-xl p-6 shadow-sm" role="alert">
				<p className="text-secondary-label">
					No feed items available at the moment. Please check back later.
				</p>
			</div>
		);
	}

	return (
		<div className="w-full">
			{feed.items.map((item, index) => (
				<FeedItem
					key={item.guid || `item-${index}`}
					item={item}
					index={index}
				/>
			))}
		</div>
	);
}
