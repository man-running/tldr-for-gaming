import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

interface NotFoundProps {
	title?: string;
	message?: string;
	context?: string;
	showBackButton?: boolean;
}

export default function NotFound({
	title = "404",
	message = "Page Not Found",
	context = "The page you are looking for doesn't exist or has been moved.",
	showBackButton = true,
}: NotFoundProps) {
	return (
		<div className="min-h-screen grid place-items-center p-4">
			<div className="max-w-sm mx-auto text-center">
				{/* Minimal number display */}
				<div className="mb-8">
					<h1 className="font-lato font-light text-6xl text-foreground/20 tracking-[2.5%]">
						{title}
					</h1>
				</div>

				{/* Clean typography hierarchy */}
				<div className="space-y-3 mb-8">
					<h2 className="font-lato font-extrabold text-[28px] leading-[125%] tracking-[2.5%] text-foreground">
						{message}
					</h2>
					<p className="font-lato font-normal text-[16px] leading-[150%] tracking-[2.5%] text-muted-foreground">
						{context}
					</p>
				</div>

				{/* Single action button */}
				{showBackButton && (
					<Button
						asChild
						className="bg-accent hover:bg-accent/90 text-accent-foreground"
					>
						<Link href="/">
							<ArrowLeft className="h-4 w-4" />
							Return Home
						</Link>
					</Button>
				)}
			</div>
		</div>
	);
}
