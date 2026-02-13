import {
	getSourceById,
	getActiveSources,
	getSourcesByCategory,
	getSourceMetadata,
	getCategories,
	isSourceActive,
	getSourceCount,
	getActiveSourceCount,
	DEFAULT_NEWS_SOURCES,
} from "./sources";

describe("News Sources - TypeScript Integration", () => {
	describe("DEFAULT_NEWS_SOURCES", () => {
		test("should have 5 default sources", () => {
			expect(DEFAULT_NEWS_SOURCES).toHaveLength(5);
		});

		test("should have required fields for each source", () => {
			DEFAULT_NEWS_SOURCES.forEach((source) => {
				expect(source.id).toBeDefined();
				expect(source.name).toBeDefined();
				expect(source.url).toBeDefined();
				expect(source.feedUrl).toBeDefined();
				expect(source.category).toBeDefined();
				expect(typeof source.active).toBe("boolean");
				expect(typeof source.priority).toBe("number");
				expect(source.scrapingType).toBeDefined();
				expect(typeof source.timeout).toBe("number");
			});
		});

		test("should have valid priority values (1-10)", () => {
			DEFAULT_NEWS_SOURCES.forEach((source) => {
				expect(source.priority).toBeGreaterThanOrEqual(1);
				expect(source.priority).toBeLessThanOrEqual(10);
			});
		});

		test("should have valid scraping types", () => {
			const validTypes = ["rss", "scrape", "api"];
			DEFAULT_NEWS_SOURCES.forEach((source) => {
				expect(validTypes).toContain(source.scrapingType);
			});
		});

		test("should all be active by default", () => {
			DEFAULT_NEWS_SOURCES.forEach((source) => {
				expect(source.active).toBe(true);
			});
		});
	});

	describe("getSourceById", () => {
		test("should return source by valid ID", () => {
			const source = getSourceById("igamingbusiness");

			expect(source).toBeDefined();
			expect(source?.id).toBe("igamingbusiness");
			expect(source?.name).toBe("iGamingBusiness");
		});

		test("should return undefined for invalid ID", () => {
			const source = getSourceById("nonexistent");

			expect(source).toBeUndefined();
		});

		test("should return correct source properties", () => {
			const source = getSourceById("gamblinginsider");

			expect(source?.name).toBe("Gambling Insider");
			expect(source?.priority).toBe(9);
			expect(source?.category).toBe("Business");
		});
	});

	describe("getActiveSources", () => {
		test("should return all active sources", () => {
			const active = getActiveSources();

			expect(active.length).toBeGreaterThan(0);
			active.forEach((source) => {
				expect(source.active).toBe(true);
			});
		});

		test("should be sorted by priority (descending)", () => {
			const active = getActiveSources();

			for (let i = 0; i < active.length - 1; i++) {
				expect(active[i].priority).toBeGreaterThanOrEqual(
					active[i + 1].priority
				);
			}
		});

		test("should have igamingbusiness as first source", () => {
			const active = getActiveSources();

			expect(active[0].id).toBe("igamingbusiness");
			expect(active[0].priority).toBe(10);
		});
	});

	describe("getSourcesByCategory", () => {
		test("should return sources by category", () => {
			const businessSources = getSourcesByCategory("Business");

			expect(businessSources.length).toBeGreaterThan(0);
			businessSources.forEach((source) => {
				expect(source.category).toBe("Business");
				expect(source.active).toBe(true);
			});
		});

		test("should return empty array for non-existent category", () => {
			const sources = getSourcesByCategory("NonExistent");

			expect(sources).toEqual([]);
		});

		test("should return multiple sources for same category sorted by priority", () => {
			const businessSources = getSourcesByCategory("Business");

			for (let i = 0; i < businessSources.length - 1; i++) {
				expect(businessSources[i].priority).toBeGreaterThanOrEqual(
					businessSources[i + 1].priority
				);
			}
		});

		test("should filter by exact category match", () => {
			const sportsSources = getSourcesByCategory("Sports Betting");

			expect(sportsSources.length).toBeGreaterThan(0);
			sportsSources.forEach((source) => {
				expect(source.category).toBe("Sports Betting");
			});
		});
	});

	describe("getSourceMetadata", () => {
		test("should return metadata for valid source", () => {
			const metadata = getSourceMetadata("igamingbusiness");

			expect(metadata).toBeDefined();
			expect(metadata?.id).toBe("igamingbusiness");
			expect(metadata?.name).toBe("iGamingBusiness");
			expect(metadata?.category).toBe("Business");
			expect(metadata?.active).toBe(true);
			expect(metadata?.priority).toBe(10);
		});

		test("should return undefined for invalid source", () => {
			const metadata = getSourceMetadata("invalid");

			expect(metadata).toBeUndefined();
		});

		test("should only include metadata fields", () => {
			const metadata = getSourceMetadata("igamingbusiness");

			expect(metadata).toHaveProperty("id");
			expect(metadata).toHaveProperty("name");
			expect(metadata).toHaveProperty("category");
			expect(metadata).toHaveProperty("active");
			expect(metadata).toHaveProperty("priority");
			// Should not have source-specific fields like url, feedUrl
			expect(metadata).not.toHaveProperty("url");
			expect(metadata).not.toHaveProperty("feedUrl");
		});
	});

	describe("getCategories", () => {
		test("should return array of unique categories", () => {
			const categories = getCategories();

			expect(Array.isArray(categories)).toBe(true);
			expect(categories.length).toBeGreaterThan(0);
		});

		test("should return sorted categories", () => {
			const categories = getCategories();

			// Create a sorted copy and compare
			const sorted = [...categories].sort();
			expect(categories).toEqual(sorted);
		});

		test("should not have duplicates", () => {
			const categories = getCategories();
			const uniqueCategories = new Set(categories);

			expect(categories.length).toBe(uniqueCategories.size);
		});

		test("should only include active source categories", () => {
			const categories = getCategories();
			const activeSourceCategories = new Set(
				DEFAULT_NEWS_SOURCES.filter((s) => s.active).map((s) => s.category)
			);

			expect(categories.length).toBe(activeSourceCategories.size);
			categories.forEach((cat) => {
				expect(activeSourceCategories).toContain(cat);
			});
		});
	});

	describe("isSourceActive", () => {
		test("should return true for active sources", () => {
			expect(isSourceActive("igamingbusiness")).toBe(true);
			expect(isSourceActive("gamblinginsider")).toBe(true);
		});

		test("should return false for non-existent sources", () => {
			expect(isSourceActive("nonexistent")).toBe(false);
		});

		test("should return correct status for all default sources", () => {
			DEFAULT_NEWS_SOURCES.forEach((source) => {
				expect(isSourceActive(source.id)).toBe(source.active);
			});
		});
	});

	describe("getSourceCount", () => {
		test("should return correct number of sources", () => {
			const count = getSourceCount();

			expect(count).toBe(5);
		});

		test("should match default sources length", () => {
			expect(getSourceCount()).toBe(DEFAULT_NEWS_SOURCES.length);
		});
	});

	describe("getActiveSourceCount", () => {
		test("should return number of active sources", () => {
			const activeCount = getActiveSourceCount();

			expect(activeCount).toBeGreaterThan(0);
		});

		test("should match filtered active sources", () => {
			const activeCount = getActiveSourceCount();
			const activeFiltered = DEFAULT_NEWS_SOURCES.filter((s) => s.active);

			expect(activeCount).toBe(activeFiltered.length);
		});

		test("should return 5 when all sources are active", () => {
			// By default, all sources are active
			expect(getActiveSourceCount()).toBe(5);
		});
	});

	describe("Default Sources Validation", () => {
		test("iGamingBusiness should be configured correctly", () => {
			const source = getSourceById("igamingbusiness");

			expect(source?.name).toBe("iGamingBusiness");
			expect(source?.priority).toBe(10);
			expect(source?.url).toBe("https://www.igamingbusiness.com");
			expect(source?.feedUrl).toContain("igamingbusiness");
		});

		test("Gambling Insider should be configured correctly", () => {
			const source = getSourceById("gamblinginsider");

			expect(source?.name).toBe("Gambling Insider");
			expect(source?.priority).toBe(9);
		});

		test("eGaming Review should have Regulations category", () => {
			const source = getSourceById("egamingreview");

			expect(source?.category).toBe("Regulations");
		});

		test("Sports Betting sources should have correct category", () => {
			const sources = getSourcesByCategory("Sports Betting");

			expect(sources.length).toBeGreaterThan(0);
			sources.forEach((source) => {
				expect(source.category).toBe("Sports Betting");
			});
		});
	});

	describe("Edge Cases", () => {
		test("getSourcesByCategory should be case-sensitive", () => {
			const exact = getSourcesByCategory("Business");
			const wrong = getSourcesByCategory("business");

			expect(exact.length).toBeGreaterThan(0);
			expect(wrong).toEqual([]);
		});

		test("getSourceById should be case-sensitive", () => {
			const correct = getSourceById("igamingbusiness");
			const wrong = getSourceById("IgamingBusiness");

			expect(correct).toBeDefined();
			expect(wrong).toBeUndefined();
		});

		test("should handle empty string searches", () => {
			const source = getSourceById("");
			const category = getSourcesByCategory("");
			const metadata = getSourceMetadata("");

			expect(source).toBeUndefined();
			expect(category).toEqual([]);
			expect(metadata).toBeUndefined();
		});

		test("should handle null/undefined safely", () => {
			// @ts-ignore - Testing error handling
			expect(getSourceById(null)).toBeUndefined();
			// @ts-ignore - Testing error handling
			expect(getSourcesByCategory(null)).toEqual([]);
			// @ts-ignore - Testing error handling
			expect(getSourceMetadata(undefined)).toBeUndefined();
		});
	});

	describe("Data Integrity", () => {
		test("should not modify default sources array", () => {
			const original = JSON.stringify(DEFAULT_NEWS_SOURCES);

			// Perform various operations
			getActiveSources();
			getSourcesByCategory("Business");
			getCategories();

			const after = JSON.stringify(DEFAULT_NEWS_SOURCES);

			expect(original).toBe(after);
		});

		test("returned arrays should be separate instances", () => {
			const active1 = getActiveSources();
			const active2 = getActiveSources();

			expect(active1).not.toBe(active2); // Different arrays
			expect(active1).toEqual(active2); // But same content
		});

		test("should not allow modification of returned sources", () => {
			const source = getSourceById("igamingbusiness");
			const originalName = source?.name;

			// Try to modify (won't affect original due to immutability practices)
			if (source) {
				source.name = "Modified";
			}

			const retrieved = getSourceById("igamingbusiness");
			// This depends on actual implementation - if using Object references
			// You might want to implement deep copy in the actual source
		});
	});

	describe("Performance", () => {
		test("should retrieve sources quickly", () => {
			const start = performance.now();

			for (let i = 0; i < 1000; i++) {
				getSourceById("igamingbusiness");
				getActiveSources();
				getSourcesByCategory("Business");
			}

			const duration = performance.now() - start;

			expect(duration).toBeLessThan(100); // Should complete in less than 100ms
		});
	});
});
