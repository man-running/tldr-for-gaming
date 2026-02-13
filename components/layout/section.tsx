"use client";

import type React from "react";
import { cn } from "@/lib/utils";

type SectionProps = {
	children: React.ReactNode;
	className?: string;
};

export function Section({ children, className }: SectionProps) {
	return (
		<div className={cn("container relative z-10", className)}>{children}</div>
	);
}
