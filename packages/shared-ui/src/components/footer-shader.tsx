"use client";

import { LiquidMetal } from "@paper-design/shaders-react";
import { DesktopOnly } from "./desktop-only";

export interface FooterShaderProps {
	/**
	 * Path to the SVG image to use for the shader effect
	 * @default "/takara.svg"
	 */
	image?: string;
	/**
	 * Background color (hex with alpha)
	 * @default "#21212100"
	 */
	colorBack?: string;
	/**
	 * Tint color (hex with alpha)
	 * @default "#ff2a007a"
	 */
	colorTint?: string;
	/**
	 * Additional className for the container
	 */
	className?: string;
	/**
	 * Additional className for the wrapper div
	 */
	wrapperClassName?: string;
}

export function FooterShader({
	image = "/takara.svg",
	colorBack = "#21212100",
	colorTint = "#ff2a007a",
	className,
	wrapperClassName,
}: FooterShaderProps) {
	return (
		<DesktopOnly className={wrapperClassName}>
			<div className="breakout-full-width px-6">
				<LiquidMetal
					image={image}
					colorBack={colorBack}
					colorTint={colorTint}
					shape={undefined}
					repetition={1.28}
					softness={1}
					shiftRed={0.96}
					shiftBlue={-0.48}
					distortion={0.93}
					contour={0.57}
					scale={1}
					angle={45}
					speed={0.3}
					offsetY={0.2}
					fit="cover"
					style={{ aspectRatio: "110 / 28" }}
					className={`w-full max-w-7xl mx-auto opacity-100 ${className || ""}`}
				/>
			</div>
		</DesktopOnly>
	);
}
