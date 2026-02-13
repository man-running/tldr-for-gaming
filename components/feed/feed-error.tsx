"use client";

import { Button } from "@/components/ui/button";

interface FeedErrorProps {
	onRetry?: () => void;
}

export function FeedError({ onRetry }: FeedErrorProps) {
	return (
		<div
			className="bg-card border border-border rounded-lg p-6"
			role="alert"
			aria-live="assertive"
		>
			<div className="space-y-3 mb-6">
				<h2 className="text-2xl font-bold text-foreground">
					Unable to load feed
				</h2>
				<p className="text-muted-foreground">
					We encountered an error while loading the feed. Please try again later
					or contact support if the problem persists.
				</p>
			</div>
			<div className="flex flex-col sm:flex-row gap-3">
				<Button onClick={() => window.location.reload()} type="button">
					Reload Page
				</Button>
				{onRetry && (
					<Button onClick={onRetry} type="button" variant="outline">
						Retry
					</Button>
				)}
			</div>
		</div>
	);
}
