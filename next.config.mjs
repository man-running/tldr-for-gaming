let userConfig;

/** @type {import('next').NextConfig} */
const withBundleAnalyzer = (await import("@next/bundle-analyzer")).default({
	enabled: process.env.ANALYZE === "true",
});

const withPWA = (await import("@ducanh2912/next-pwa")).default({
	dest: "public",
	register: true,
	skipWaiting: true,
	cacheOnFrontEndNav: true,
	disable: true,
	clientsClaim: true,
	runtimeCaching: [
		{
			urlPattern: /^https:\/\/challenges\.cloudflare\.com\/.*/i,
			handler: "NetworkOnly",
		},
		{
			urlPattern: /^https:\/\/cdnjs\.cloudflare\.com\/.*/i,
			handler: "NetworkOnly",
		},
	],
	extendDefaultRuntimeCaching: true,
});

const nextConfig = withPWA(
	withBundleAnalyzer({
		transpilePackages: ["@takara/shared-ui"],
		images: {
			unoptimized: true,
			remotePatterns: [
				{
					protocol: "https",
					hostname: "tldr.takara.ai",
				},
				{
					protocol: "http",
					hostname: "localhost",
				},
				// Allow images from any domain (for news articles)
				{
					protocol: "https",
					hostname: "**",
				},
				{
					protocol: "http",
					hostname: "**",
				},
			],
		},
		typedRoutes: true,
		eslint: {
			ignoreDuringBuilds: true,
		},
		experimental: {
			webpackBuildWorker: true,
			parallelServerBuildTraces: true,
			parallelServerCompiles: true,
			optimizePackageImports: ["lucide-react", "@radix-ui/react-icons"],
		},
		async headers() {
			return [
				{
					source: "/:path*",
					headers: [
						{ key: "X-DNS-Prefetch-Control", value: "on" },
						{ key: "X-XSS-Protection", value: "1; mode=block" },
						{ key: "X-Frame-Options", value: "SAMEORIGIN" },
						{ key: "X-Content-Type-Options", value: "nosniff" },
						{
							key: "Referrer-Policy",
							value: "strict-origin-when-cross-origin",
						},
						{
							key: "Permissions-Policy",
							value:
								"camera=(), microphone=(), geolocation=(), interest-cohort=()",
						},
					],
				},
				{
					source: "/sw.js",
					headers: [
						{
							key: "Content-Type",
							value: "application/javascript; charset=utf-8",
						},
						{
							key: "Cache-Control",
							value: "no-cache, no-store, must-revalidate",
						},
						{
							key: "Content-Security-Policy",
							value: "default-src 'self'; script-src 'self'",
						},
					],
				},
				{
					source: "/hf/papers",
					headers: [
						{
							key: "CDN-Cache-Control",
							value: "max-age=300, stale-while-revalidate=600",
						},
					],
				},
			];
		},
		async rewrites() {
			return [
				{
					source: "/ingest/static/:path*",
					destination: "https://eu-assets.i.posthog.com/static/:path*",
				},
				{
					source: "/ingest/:path*",
					destination: "https://eu.i.posthog.com/:path*",
				},
				{
					source: "/hf/papers",
					destination: "https://huggingface.co/api/papers/search",
				},
			];
		},
		skipTrailingSlashRedirect: true,
	}),
);

mergeConfig(nextConfig, userConfig);

function mergeConfig(nextConfig, userConfig) {
	if (!userConfig) {
		return;
	}

	for (const key in userConfig) {
		if (
			typeof nextConfig[key] === "object" &&
			!Array.isArray(nextConfig[key])
		) {
			nextConfig[key] = {
				...nextConfig[key],
				...userConfig[key],
			};
		} else {
			nextConfig[key] = userConfig[key];
		}
	}
}

export default nextConfig;
