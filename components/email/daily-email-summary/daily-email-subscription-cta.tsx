"use client";

import { type ChangeEvent, type FormEvent, useId, useState } from "react";
import Turnstile from "react-turnstile";
import { logger } from "@/lib/logger";

type StatusType = {
	type: "error" | "success" | "loading";
	message: string;
} | null;

export function DailyEmailSubscriptionCTA() {
	const [email, setEmail] = useState<string>("");
	const [status, setStatus] = useState<StatusType>(null);
	const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
	const [turnstileToken, setTurnstileToken] = useState<string>("");
	const inputId = useId();

	const validateEmail = (email: string): boolean => {
		return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
	};

	const handleSubmit = async (e: FormEvent<HTMLFormElement>): Promise<void> => {
		e.preventDefault();

		if (!validateEmail(email)) {
			setStatus({ type: "error", message: "Invalid email address" });
			return;
		}

		if (!turnstileToken) {
			setStatus({ type: "error", message: "Please complete the verification" });
			return;
		}

		setIsSubmitting(true);
		setStatus({ type: "loading", message: "Subscribing..." });

		// API CALL
		try {
			const response = await fetch("/api/subscribe", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ email, turnstileToken }),
			});

			if (response.ok) {
				setStatus({
					type: "success",
					message: "You've been subscribed successfully!",
				});
				setEmail(""); // Clear input
				setTurnstileToken("");
			} else {
				const errorData = await response.json();
				setStatus({
					type: "error",
					message: errorData.error || "Subscription failed. Try again.",
				});
				setTurnstileToken("");
			}
		} catch (error) {
			logger.error(
				"Error subscribing",
				error instanceof Error ? error : undefined,
			);
			setStatus({
				type: "error",
				message: "An unexpected error occurred. Try again.",
			});
			setTurnstileToken("");
		} finally {
			setIsSubmitting(false);
		}
	};

	const handleEmailChange = (e: ChangeEvent<HTMLInputElement>): void => {
		setEmail(e.target.value);
	};

	const handleTurnstileVerify = (token: string) => {
		setTurnstileToken(token);
	};

	const handleTurnstileError = () => {
		setStatus({
			type: "error",
			message: "Verification failed. Please try again.",
		});
		setTurnstileToken("");
	};

	const handleTurnstileExpire = () => {
		setTurnstileToken("");
	};

	return (
		<div className="w-full mx-auto">
			<h2 className="font-lato font-extrabold text-[35px] sm:text-[40px] leading-[150%] sm:leading-[125%] tracking-[2.5%] align-middle sm:align-baseline text-foreground mb-[25px]">
				Get tldr.<span className="text-accent">takara.ai</span> to Your Email,
				Everyday.
			</h2>
			<form
				onSubmit={handleSubmit}
				className="space-y-4"
				data-ph-capture-attribute-form-type="email-subscription"
			>
				<div>
					<div className="flex flex-col sm:flex-row gap-[25px]">
						<input
							className="form-input flex-1"
							id={inputId}
							type="email"
							value={email}
							onChange={handleEmailChange}
							placeholder="yourname@example.com"
							disabled={isSubmitting}
							required
						/>
						<button
							className={`px-6 py-2 rounded-md font-medium transition-colors whitespace-nowrap cursor-pointer ${
								isSubmitting || !email
									? "bg-gray-300 text-gray-500 cursor-not-allowed"
									: "btn-primary"
							}`}
							type="submit"
							disabled={isSubmitting || !email}
							data-ph-capture-attribute-action="subscribe"
						>
							{isSubmitting ? "Subscribing..." : "Subscribe"}
						</button>
					</div>
				</div>

				{/* Turnstile Widget */}
				<div className="flex justify-center">
					<Turnstile
						sitekey={process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY ?? ""}
						onVerify={handleTurnstileVerify}
						onError={handleTurnstileError}
						onExpire={handleTurnstileExpire}
						theme="light"
						size="normal"
					/>
				</div>

				{status && (
					<div
						className={`font-lato font-normal text-sm px-3 py-2 rounded-md ${
							status.type === "error"
								? "status-error"
								: status.type === "success"
									? "status-success"
									: "status-loading"
						}`}
					>
						{status.message}
					</div>
				)}
			</form>
			<p className="font-lato font-normal text-xs leading-[125%] tracking-[2.5%] text-muted-foreground mt-4">
				We respect your privacy. Unsubscribe at any time.
			</p>
		</div>
	);
}
