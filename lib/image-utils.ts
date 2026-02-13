import { createHash } from "crypto";

const BASE62_CHARS =
	"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";

const VERCEL_BLOB_BASE_URL =
	"https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com";

/**
 * Encodes a Buffer to base62 string
 * This matches the Go implementation exactly for perfect parity
 */
function encodeBase62(data: Buffer): string {
	if (data.length === 0) {
		return "0";
	}

	// Convert buffer to BigInt (treating as big-endian)
	let num = BigInt(0);
	for (let i = 0; i < data.length; i++) {
		num = num * BigInt(256) + BigInt(data[i]!);
	}

	// Convert to base62
	const base = BigInt(62);
	const zero = BigInt(0);
	let result = "";
	let n = num;

	if (n === zero) {
		return "0";
	}

	while (n > zero) {
		const remainder = n % base;
		result = BASE62_CHARS[Number(remainder)]! + result;
		n = n / base; // BigInt division truncates (integer division)
	}

	return result;
}

/**
 * Generates a deterministic blob key from input string using SHA256 and base62 encoding
 * This matches the Go implementation exactly for perfect parity
 */
export function generateBlobKey(input: string): string {
	const hash = createHash("sha256").update(input).digest();
	return encodeBase62(hash);
}

/**
 * Generates the full blob path for a spectrogram image
 */
export function generateSpectrogramBlobPath(paperTitle: string): string {
	const key = generateBlobKey(paperTitle);
	return `media/papers/spectrogram/${key}.webp`;
}

/**
 * Constructs the deterministic blob URL for a spectrogram image
 */
export function constructSpectrogramBlobURL(paperTitle: string): string {
	const path = generateSpectrogramBlobPath(paperTitle);
	return `${VERCEL_BLOB_BASE_URL}/${path}`;
}

/**
 * Fetches an image with fallback: tries blob URL first, falls back to API if 404
 * Returns the final image URL
 */
export async function fetchImageWithFallback(
	blobURL: string,
	apiURL: string,
): Promise<string> {
	// Try fetching from blob first using HEAD to check existence
	try {
		const response = await fetch(blobURL, {
			method: "HEAD",
			next: { revalidate: 86400 }, // Cache for 24 hours
		});

		if (response.ok) {
			// Blob exists, return the blob URL
			return blobURL;
		}

		// If 404, blob doesn't exist, need to generate via API
		if (response.status === 404) {
			// Call API to generate and store image
			const apiResponse = await fetch(apiURL, {
				next: { revalidate: 86400 },
			});
			if (!apiResponse.ok) {
				throw new Error(
					`API failed to generate image: ${apiResponse.status} ${apiResponse.statusText}`,
				);
			}

			const data = (await apiResponse.json()) as { url?: string };
			if (data.url) {
				return data.url;
			}

			throw new Error("API response missing url field");
		}

		// Other error status from blob, try API as fallback
		const apiResponse = await fetch(apiURL, {
			next: { revalidate: 86400 },
		});
		if (!apiResponse.ok) {
			throw new Error(
				`Blob fetch failed (${response.status}) and API also failed: ${apiResponse.status} ${apiResponse.statusText}`,
			);
		}

		const data = (await apiResponse.json()) as { url?: string };
		if (data.url) {
			return data.url;
		}

		throw new Error("API response missing url field");
	} catch (error) {
		// If fetch fails (network error, etc.), try API as fallback
		if (
			error instanceof TypeError ||
			(error instanceof Error && !error.message.includes("API"))
		) {
			const apiResponse = await fetch(apiURL, {
				next: { revalidate: 86400 },
			});
			if (!apiResponse.ok) {
				throw new Error(
					`Network error and API also failed: ${apiResponse.status} ${apiResponse.statusText}`,
				);
			}

			const data = (await apiResponse.json()) as { url?: string };
			if (data.url) {
				return data.url;
			}

			throw new Error("API response missing url field");
		}

		// Re-throw if it's already an API error
		throw error;
	}
}
