import type { ReactNode } from "react";
import { cn } from "../utils";

interface HeroContainerProps {
	children: ReactNode;
	className?: string;
}

export function HeroContainer({ children, className }: HeroContainerProps) {
	return (
		<div
			className={cn(
				"mx-auto max-w-5xl px-6 2xl:pr-[10%] lg:pr-[30%]",
				className,
			)}
		>
			{children}
		</div>
	);
}
