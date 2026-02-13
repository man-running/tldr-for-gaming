import type { MetadataRoute } from "next";
import { siteConfig } from "@/lib/constants";

export default function manifest(): MetadataRoute.Manifest {
	return {
		name: siteConfig.name,
		short_name: "TakaraTLDR",
		description: siteConfig.description,
		start_url: "/",
		display: "standalone",
		background_color: "#ffffff",
		theme_color: "#D91009",
		icons: [
			{
				src: "/web-app-manifest-192x192.png",
				sizes: "192x192",
				type: "image/png",
				purpose: "any",
			},
			{
				src: "/web-app-manifest-512x512.png",
				sizes: "512x512",
				type: "image/png",
				purpose: "maskable",
			},
		],
	};
}
