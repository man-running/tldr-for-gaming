"use client";

import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "../utils";

export interface NavbarLayoutProps
	extends Omit<HTMLAttributes<HTMLDivElement>, "children"> {
	children: ReactNode;
}

export function NavbarLayout({
	className,
	children,
	...props
}: NavbarLayoutProps) {
	return (
		<div
			className={cn("mx-auto flex items-center gap-2 w-full", className)}
			{...props}
		>
			{children}
		</div>
	);
}
