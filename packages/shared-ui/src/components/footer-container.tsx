import type { ReactNode } from "react";
import { Ray } from "./ray";

interface FooterContainerProps {
	children: ReactNode;
}

export function FooterContainer({ children }: FooterContainerProps) {
	return (
		<div className="bg-linear-to-t dark to-background from-[#050505] text-foreground overflow-hidden relative mask-zigzag-top">
			<Ray className="absolute left-1/4 top-0 size-160 text-accent fill-accent blur-2xl rotate-45 -translate-x-1/2 -translate-y-1/2 opacity-80" />
			<Ray className="absolute left-4/5 -top-20 size-160 text-accent fill-accent blur-2xl rotate-45 -translate-x-1/2 -translate-y-1/2 opacity-70" />
			<Ray className="absolute left-3/4 top-0 size-160 text-accent fill-accent blur-2xl rotate-45 -translate-x-1/2 -translate-y-1/2 opacity-80" />
			<Ray className="absolute left-2/5 top-0 size-300 opacity-50 text-accent fill-accent blur-2xl rotate-45 -translate-x-1/2 -translate-y-1/2" />
			<div className="max-w-5xl mx-auto px-6 pt-20">{children}</div>
		</div>
	);
}
