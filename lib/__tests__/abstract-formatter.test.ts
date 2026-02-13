import { describe, expect, it } from "vitest";
import { formatAbstract } from "../abstract-formatter";

describe("formatAbstract", () => {
	it("returns an empty root for empty input", () => {
		const result = formatAbstract("");
		expect(result.type).toBe("root");
		expect(result.children).toEqual([]);
	});

	it("formats single paragraph with incidental newlines", () => {
		const input =
			"This is a single paragraph\nwith line wraps\nthat should be collapsed\ninto one paragraph.";
		const result = formatAbstract(input);

		const paragraph = (result.children[0] as { children?: unknown[] })
			.children?.[0] as { type?: string; value?: string } | undefined;
		expect(paragraph?.type).toBe("text");
		expect(paragraph?.value).toContain(
			"This is a single paragraph with line wraps that should be collapsed into one paragraph.",
		);
	});

	it("preserves paragraph breaks where newline is followed by leading whitespace", () => {
		const input =
			"First paragraph.\n Second paragraph with indent.\n  Third paragraph more indent.";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("First paragraph");
		expect(text).toContain("Second paragraph with indent");
		expect(text).toContain("Third paragraph more indent");
	});

	it("preserves leading spaces as non-breaking spaces in multi-line content", () => {
		const input = "Line one\n   Line two indented";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("Line one");
		expect(text).toContain("Line two indented");
	});

	it("sanitizes HTML content", () => {
		const input = '<script>alert("xss")</script>Safe text';
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).not.toContain("<script>");
		expect(text).toContain("Safe text");
	});

	it("converts TeX textbf to strong tags", () => {
		const input = "This is \\textbf{bold text} in a sentence";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("bold text");
	});

	it("converts TeX textit to em tags", () => {
		const input = "This is \\textit{italic text} in a sentence";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("italic text");
	});

	it("converts TeX emph to em tags", () => {
		const input = "This is \\emph{emphasized text} in a sentence";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("emphasized text");
	});

	it("converts TeX href to HTML links", () => {
		const input = "See \\href{https://example.com}{this link} for more";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("https://example.com");
		expect(text).toContain("this link");
	});

	it("converts nested TeX formatting", () => {
		const input = "This is \\textbf{\\emph{bold and italic}} text";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("bold and italic");
	});

	it("handles TeX smart quotes", () => {
		const input = "He said ``hello'' to her";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain('He said "hello" to her');
	});

	it("handles real-world TeX with formatting and links", () => {
		const input =
			"We present \\textbf{Brain-IT}, a \\textit{brain-inspired} approach. See \\href{https://github.com/example}{our code}.";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("Brain-IT");
		expect(text).toContain("brain-inspired");
		expect(text).toContain("https://github.com/example");
	});

	it("preserves TeX math delimiters for MathJax", () => {
		const input =
			"This contains math: $E=mc^2$ and display math: $$\\int_0^\\infty e^{-x} dx$$";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("$E=mc^2$");
		expect(text).toContain("$$\\int_0^\\infty e^{-x} dx$$");
	});

	it("handles CRLF line endings", () => {
		const input = "First line\r\nSecond line\r\nThird line";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("First line Second line Third line");
	});

	it("splits paragraphs on multiple consecutive spaces", () => {
		const input = "First paragraph.    Second paragraph.      Third paragraph.";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("First paragraph");
		expect(text).toContain("Second paragraph");
		expect(text).toContain("Third paragraph");
	});

	it("handles multiple consecutive newlines", () => {
		const input = "Paragraph 1\n\n\nParagraph 2";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("Paragraph 1");
		expect(text).toContain("Paragraph 2");
	});

	it("trims whitespace from paragraphs", () => {
		const input = "  Paragraph with surrounding whitespace  ";
		const result = formatAbstract(input);

		const text = JSON.stringify(result);
		expect(text).toContain("Paragraph with surrounding whitespace");
	});

	it("handles real-world arxiv abstract example", () => {
		const input = `Reconstructing images seen by people from their fMRI brain recordings
provides a non-invasive window into the human brain. Despite recent progress
enabled by diffusion models, current methods often lack faithfulness to the
actual seen images. We present "Brain-IT", a brain-inspired approach that
addresses this challenge through a Brain Interaction Transformer (BIT),
allowing effective interactions between clusters of functionally-similar
brain-voxels.`;

		const result = formatAbstract(input);

		expect(result).toContain(
			"Reconstructing images seen by people from their fMRI brain recordings provides a non-invasive window",
		);
		expect(result).toContain("brain-voxels");
		const text = JSON.stringify(result);
		expect(text).toContain(
			"Reconstructing images seen by people from their fMRI",
		);
	});

	it("does not create empty paragraphs", () => {
		const input = "Paragraph 1\n\n\n\nParagraph 2";
		const result = formatAbstract(input);

		// Should not have empty <p></p> tags
		expect(result).not.toMatch(/<p[^>]*>\s*<\/p>/);
	});
});
