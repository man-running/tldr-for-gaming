"use client";

import Link from "next/link";
import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { logger } from "@/lib/logger";

/**
 * ErrorBoundary component for handling application errors.
 * @param error - The error object thrown by the application.
 * @param reset - Function to reset the error boundary state.
 */
export default function ErrorBoundary({
	error,
	reset,
}: {
	error: Error & { digest?: string };
	reset: () => void;
}) {
	useEffect(() => {
		// Log the error to an error reporting service
		logger.error("Application error", error);
	}, [error]);

	return (
		<div className="min-h-screen grid place-items-center p-4">
			<div className="max-w-md w-full text-center" role="alert">
				<div className="space-y-4 mb-8">
					<h2 className="text-4xl lg:text-5xl font-bold leading-tight tracking-tight text-foreground">
						Something went wrong
					</h2>
					<p className="text-muted-foreground">
						We apologize for the inconvenience. An unexpected error has
						occurred.
					</p>
				</div>
				<div className="flex flex-col sm:flex-row gap-3 justify-center">
					<Button type="button" onClick={() => reset()}>
						Try again
					</Button>
					<Button asChild variant="outline">
						<Link href="/">Go to homepage</Link>
					</Button>
				</div>
			</div>
		</div>
	);
}
