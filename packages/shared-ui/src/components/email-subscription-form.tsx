"use client";

import { CheckIcon } from "lucide-react";
import { type FormEvent, useState } from "react";
import { Button } from "./button";
import { Input } from "./input";

let Turnstile: typeof import("react-turnstile").default | null = null;
try {
	Turnstile = require("react-turnstile").default;
} catch {
	// Turnstile is optional
}

interface EmailSubscriptionFormProps {
	onSubmit?: (email: string) => Promise<void>;
	endpoint?: string;
	turnstileSiteKey?: string;
}

export function EmailSubscriptionForm({
	onSubmit,
	endpoint,
	turnstileSiteKey,
}: EmailSubscriptionFormProps) {
	const [isSubmitting, setIsSubmitting] = useState(false);
	const [isSuccess, setIsSuccess] = useState(false);
	const [turnstileToken, setTurnstileToken] = useState<string>("");

	async function handleSubmit(e: FormEvent<HTMLFormElement>) {
		e.preventDefault();
		setIsSubmitting(true);
		setIsSuccess(false);
		const formData = new FormData(e.currentTarget);
		const email = formData.get("email") as string;

		if (!email) {
			alert("Please enter a valid email address.");
			setIsSubmitting(false);
			return;
		}

		if (endpoint && turnstileSiteKey && !turnstileToken) {
			alert("Please complete the verification.");
			setIsSubmitting(false);
			return;
		}

		try {
			if (endpoint) {
				const body: { email: string; turnstileToken?: string } = { email };
				if (turnstileSiteKey && turnstileToken) {
					body.turnstileToken = turnstileToken;
				}

				const response = await fetch(endpoint, {
					method: "POST",
					headers: {
						"Content-Type": "application/json",
					},
					body: JSON.stringify(body),
				});

				if (!response.ok) {
					const errorData = await response.json().catch(() => ({}));
					throw new Error(
						errorData.error || "Something went wrong. Please try again later.",
					);
				}
			} else if (onSubmit) {
				await onSubmit(email);
			} else {
				throw new Error("No submission handler provided");
			}
			setIsSuccess(true);
			setTurnstileToken("");
		} catch (error) {
			console.error(error);
			alert(
				error instanceof Error
					? error.message
					: "Something went wrong. Please try again later.",
			);
			setTurnstileToken("");
		} finally {
			setIsSubmitting(false);
		}
	}

	return (
		<form onSubmit={handleSubmit} className="w-fit ml-auto">
			<div className="flex gap-2 w-full sm:w-auto bg-muted-foreground/10 rounded-xl p-2 relative overflow-hidden">
				{isSuccess && (
					<div className="absolute bg-green-700 text-green-200 inset-2 rounded-sm flex items-center justify-center gap-3 text-sm font-medium">
						<CheckIcon className="size-4" />
						You&apos;re on the list!
					</div>
				)}
				<Input
					type="email"
					name="email"
					placeholder="my@email.com"
					required
					disabled={isSubmitting}
					className="flex-1 sm:flex-initial sm:w-64 border-none bg-transparent!"
				/>
				<Button type="submit" variant="default" disabled={isSubmitting}>
					{isSubmitting ? "Subscribing..." : "Subscribe"}
				</Button>
			</div>
			{turnstileSiteKey && Turnstile && (
				<div className="flex justify-center mt-2">
					<Turnstile
						sitekey={turnstileSiteKey}
						onVerify={setTurnstileToken}
						onError={() => setTurnstileToken("")}
						onExpire={() => setTurnstileToken("")}
						theme="light"
						size="normal"
					/>
				</div>
			)}
		</form>
	);
}
