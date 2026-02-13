"use client";

import { useMotionValueEvent, useScroll } from "motion/react";
import { type HTMLAttributes, type ReactNode, useState } from "react";
import { cn } from "@/lib/utils";

export function StyleOnScroll({
	top,
	notTop,
	bottom,
	notBottom,
	children,
	className,
	...props
}: {
	top?: string;
	notTop?: string;
	bottom?: string;
	notBottom?: string;
	children: ReactNode;
} & Omit<HTMLAttributes<HTMLDivElement>, "children">) {
	const { scrollYProgress } = useScroll();
	const [isTop, setIsTop] = useState(scrollYProgress.get() <= 0.01);
	const [isBottom, setIsBottom] = useState(scrollYProgress.get() >= 0.99);

	useMotionValueEvent(scrollYProgress, "change", (latest) => {
		setIsTop(latest <= 0.01);
		setIsBottom(latest >= 0.99);
	});

	return (
		<div
			className={cn(
				className,
				isTop && top,
				!isTop && notTop,
				isBottom && bottom,
				!isBottom && notBottom,
			)}
			{...props}
		>
			{children}
		</div>
	);
}
