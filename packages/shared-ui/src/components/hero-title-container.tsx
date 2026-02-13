import type { ReactNode } from "react";
import { cn } from "../utils";

interface HeroTitleContainerProps {
	children: ReactNode;
	outerClassName?: string;
	innerClassName?: string;
}

export function HeroTitleContainer({
	children,
	outerClassName,
	innerClassName,
}: HeroTitleContainerProps) {
	return (
		<div
			className={cn(
				"lg:h-screen h-[40vh] top-[50vh] right-0 lg:ml-10 xl:ml-20 2xl:ml-40 mx-6 flex absolute lg:top-0 left-0 lg:right-1/2",
				outerClassName,
			)}
		>
			<div
				className={cn(
					"flex flex-col w-full justify-center items-start py-20",
					innerClassName,
				)}
			>
				{children}
			</div>
		</div>
	);
}
