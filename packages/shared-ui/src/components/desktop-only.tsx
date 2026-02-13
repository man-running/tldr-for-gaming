"use client";

import type { ReactNode } from "react";
import { cn } from "../utils";

export function DesktopOnly({
	children,
	className,
}: {
	children: ReactNode;
	className?: string;
}) {
	return <div className={cn(className)}>{children}</div>;
}
