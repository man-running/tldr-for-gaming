"use client";

import { HeroTitleContainer } from "@takara/shared-ui/hero-title-container";
import { ChevronsDown } from "lucide-react";
import Image from "next/image";
import { StyleOnScroll } from "../style-on-scroll";
import { useFeed } from "./feed-data-provider";

function formatDate(dateString?: string): string {
	if (!dateString) return "No date available";

	try {
		return new Date(dateString).toLocaleDateString("en-US", {
			month: "long",
			day: "numeric",
			year: "numeric",
		});
	} catch {
		return "No date available";
	}
}

export function HeaderContent() {
	const { feed, loading, currentDate } = useFeed();

	const formattedDate = loading
		? "Loading..."
		: formatDate(feed?.lastBuildDate ?? currentDate ?? "");

	const summary = loading
		? "Loading daily AI research summaries..."
		: feed?.description || "Daily AI research summaries";

	return (
		<div className="w-full min-h-screen translate-y-[-68px] relative">
			<div className="lg:w-1/2 w-full ml-auto z-10 lg:h-screen h-[60vh] pointer-events-none flex items-center justify-center">
				<div className="w-full h-full relative p-10">
					<Image
						src="/icon.svg"
						alt="Origami Crane Logo"
						width={360}
						height={280}
						className="object-contain w-full h-full"
						priority
					/>
				</div>
			</div>
			<div className="w-full h-screen flex items-end justify-center absolute inset-0 bottom-auto">
				<StyleOnScroll
					top="opacity-100"
					notTop="opacity-0"
					className="flex items-center gap-2 p-8 transition-opacity duration-300"
				>
					<ChevronsDown className="size-5 text-muted-foreground" />
				</StyleOnScroll>
			</div>
			<HeroTitleContainer innerClassName="md:text-8xl text-5xl sm:text-7xl font-semibold tracking-tighter text-foreground">
				<div className="flex flex-col gap-4">
					<h1>TLDR: {formattedDate}</h1>
					<p className="text-muted-foreground text-2xl text-pretty mt-12 tracking-normal font-medium">
						{summary}
					</p>
				</div>
			</HeroTitleContainer>
		</div>
	);
}
