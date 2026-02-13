import { cn } from "../utils";

export function Ray({ className }: { className?: string }) {
	return (
		<svg
			className={cn("origin-center pointer-events-none", className)}
			width="1000"
			height="1000"
			viewBox="0 0 1000 1000"
			fill="none"
			xmlns="http://www.w3.org/2000/svg"
			role="img"
			aria-label="Ray graphic"
		>
			<path
				d="M495.882 996.38L478.558 442.275L504.119 3.6204L538.364 435.455L495.882 996.38Z"
				fill="currentColor"
			/>
		</svg>
	);
}
