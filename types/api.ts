/**
 * API Types for Takara TLDR Paper Service
 * OpenAPI 3.0.3 compatible schemas
 */

// Core paper data (common between sources)
interface CorePaperData {
	title: string;
	abstract: string;
	authors: string[];
	publishedDate?: string;
	arxivId: string;
	pdfUrl: string;
}

// HuggingFace-specific features
interface HuggingFaceFeatures {
	upvotes?: number;
	githubUrl?: string;
	collections?: Collection[];
	huggingfaceUrl?: string;
}

// ArXiv-specific features
interface ArxivFeatures {
	arxivUrl?: string;
	categories?: string[];
	doi?: string;
}

// Combined paper data structure
export interface PaperData
	extends CorePaperData,
		HuggingFaceFeatures,
		ArxivFeatures {}

// Collection structure
interface Collection {
	name: string;
	itemCount: number;
}

// Paper API response

// Error codes enum
enum ErrorCode {
	INVALID_ARXIV_ID = "INVALID_ARXIV_ID",
	PAPER_NOT_FOUND = "PAPER_NOT_FOUND",
	EXTERNAL_API_ERROR = "EXTERNAL_API_ERROR",
	RATE_LIMITED = "RATE_LIMITED",
	INTERNAL_ERROR = "INTERNAL_ERROR",
	VALIDATION_ERROR = "VALIDATION_ERROR",
}

// Cache entry interface
interface _CacheEntry<T = unknown> {
	data: T;
	timestamp: number;
	ttl: number;
}

// Data source tracking
interface _DataSourceInfo {
	source: "huggingface" | "arxiv" | "combined";
	hasHfData: boolean;
	hasArxivData: boolean;
	fallbackUsed: boolean;
}

// HuggingFace API response (internal)

// OpenAPI Schema definitions
const _OpenApiSchemas = {
	CorePaperData: {
		type: "object",
		required: ["title", "abstract", "authors", "arxivId", "pdfUrl"],
		properties: {
			title: {
				type: "string",
				description: "Paper title",
				example: "Attention Is All You Need",
			},
			abstract: {
				type: "string",
				description: "Paper abstract",
				example:
					"The dominant sequence transduction models are based on complex recurrent or convolutional neural networks...",
			},
			authors: {
				type: "array",
				items: { type: "string" },
				description: "List of author names",
				example: ["Ashish Vaswani", "Noam Shazeer", "Niki Parmar"],
			},
			publishedDate: {
				type: "string",
				format: "date",
				description: "Publication date",
				example: "2017-06-12",
			},
			arxivId: {
				type: "string",
				description: "ArXiv identifier",
				example: "1706.03762",
			},
			pdfUrl: {
				type: "string",
				format: "uri",
				description: "PDF download URL",
				example: "https://arxiv.org/pdf/1706.03762.pdf",
			},
		},
	},
	HuggingFaceFeatures: {
		type: "object",
		properties: {
			upvotes: {
				type: "integer",
				description: "Number of upvotes on HuggingFace",
				example: 1500,
			},
			githubUrl: {
				type: "string",
				format: "uri",
				description: "GitHub repository URL",
				example: "https://github.com/tensorflow/tensor2tensor",
			},
			collections: {
				type: "array",
				items: {
					type: "object",
					properties: {
						name: { type: "string" },
						itemCount: { type: "integer" },
					},
				},
				description: "Collections the paper belongs to",
			},
			huggingfaceUrl: {
				type: "string",
				format: "uri",
				description: "HuggingFace paper URL",
				example: "https://huggingface.co/papers/1706.03762",
			},
		},
	},
	ArxivFeatures: {
		type: "object",
		properties: {
			arxivUrl: {
				type: "string",
				format: "uri",
				description: "ArXiv abstract URL",
				example: "https://arxiv.org/abs/1706.03762",
			},
			categories: {
				type: "array",
				items: { type: "string" },
				description: "ArXiv categories",
				example: ["cs.CL", "cs.LG"],
			},
			doi: {
				type: "string",
				description: "Digital Object Identifier",
				example: "10.48550/arXiv.1706.03762",
			},
		},
	},
	PaperData: {
		type: "object",
		required: ["title", "abstract", "authors", "arxivId", "pdfUrl"],
		properties: {
			// Core data
			title: {
				type: "string",
				description: "Paper title",
				example: "Attention Is All You Need",
			},
			abstract: {
				type: "string",
				description: "Paper abstract",
				example:
					"The dominant sequence transduction models are based on complex recurrent or convolutional neural networks...",
			},
			authors: {
				type: "array",
				items: { type: "string" },
				description: "List of author names",
				example: ["Ashish Vaswani", "Noam Shazeer", "Niki Parmar"],
			},
			publishedDate: {
				type: "string",
				format: "date",
				description: "Publication date",
				example: "2017-06-12",
			},
			arxivId: {
				type: "string",
				description: "ArXiv identifier",
				example: "1706.03762",
			},
			pdfUrl: {
				type: "string",
				format: "uri",
				description: "PDF download URL",
				example: "https://arxiv.org/pdf/1706.03762.pdf",
			},
			// HuggingFace features
			upvotes: {
				type: "integer",
				description: "Number of upvotes on HuggingFace",
				example: 1500,
			},
			githubUrl: {
				type: "string",
				format: "uri",
				description: "GitHub repository URL",
				example: "https://github.com/tensorflow/tensor2tensor",
			},
			collections: {
				type: "array",
				items: {
					type: "object",
					properties: {
						name: { type: "string" },
						itemCount: { type: "integer" },
					},
				},
				description: "Collections the paper belongs to",
			},
			huggingfaceUrl: {
				type: "string",
				format: "uri",
				description: "HuggingFace paper URL",
				example: "https://huggingface.co/papers/1706.03762",
			},
			// ArXiv features
			arxivUrl: {
				type: "string",
				format: "uri",
				description: "ArXiv abstract URL",
				example: "https://arxiv.org/abs/1706.03762",
			},
			categories: {
				type: "array",
				items: { type: "string" },
				description: "ArXiv categories",
				example: ["cs.CL", "cs.LG"],
			},
			doi: {
				type: "string",
				description: "Digital Object Identifier",
				example: "10.48550/arXiv.1706.03762",
			},
		},
	},
	ApiError: {
		type: "object",
		required: ["code", "message"],
		properties: {
			code: {
				type: "string",
				enum: Object.values(ErrorCode),
				description: "Error code",
			},
			message: {
				type: "string",
				description: "Human-readable error message",
			},
			details: {
				type: "object",
				description: "Additional error details",
			},
		},
	},
	ApiMeta: {
		type: "object",
		properties: {
			cached: {
				type: "boolean",
				description: "Whether response was served from cache",
			},
			cacheSource: {
				type: "string",
				enum: ["memory", "stale"],
				description: "Cache source type",
			},
			cacheAge: {
				type: "integer",
				description: "Cache age in milliseconds",
			},
			dataComplete: {
				type: "boolean",
				description: "Whether all data fields are complete",
			},
			dataSource: {
				type: "string",
				enum: ["huggingface", "arxiv", "combined"],
				description: "Primary data source used",
			},
			responseTime: {
				type: "integer",
				description: "Response time in milliseconds",
			},
			timestamp: {
				type: "string",
				format: "date-time",
				description: "Response timestamp",
			},
			requestId: {
				type: "string",
				description: "Unique request identifier",
			},
		},
	},
	PaperApiResponse: {
		type: "object",
		required: ["success"],
		properties: {
			success: {
				type: "boolean",
				description: "Request success status",
			},
			data: {
				$ref: "#/components/schemas/PaperData",
			},
			error: {
				$ref: "#/components/schemas/ApiError",
			},
			meta: {
				$ref: "#/components/schemas/ApiMeta",
			},
		},
	},
};
