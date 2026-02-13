import type { Root } from "hast";
import remarkGfm from "remark-gfm";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import { unified } from "unified";

type FormattedContent = Root;

const processor = unified().use(remarkParse).use(remarkGfm).use(remarkRehype);

function texToMarkdown(text: string): string {
	let result = text;

	// Normalize whitespace: collapse multiple spaces and newlines to single space
	result = result.replace(/\s+/g, " ").trim();

	// Handle \href{url}{text} -> [text](url)
	result = result.replace(/\\href\{([^}]+)\}\{([^}]+)\}/g, "[$2]($1)");

	// Handle \textbf{text} -> **text** (including potential leading space)
	result = result.replace(/\\textbf\s*\{\s*([^}]+?)\s*\}/g, "**$1**");

	// Handle \emph{text} and \textit{text} -> *text* (including potential spaces)
	result = result.replace(/\\(?:emph|textit)\s*\{\s*([^}]+?)\s*\}/g, "*$1*");

	// Handle \texttt{text} -> `text` (including potential spaces)
	result = result.replace(/\\texttt\s*\{\s*([^}]+?)\s*\}/g, "`$1`");

	// Handle \text{text} -> just text (including potential spaces)
	result = result.replace(/\\text\s*\{\s*([^}]+?)\s*\}/g, "$1");

	// Fix smart quotes
	result = result.replace(/``/g, '"');
	result = result.replace(/''/g, '"');

	return result;
}

export function formatAbstract(raw: string): FormattedContent {
	if (!raw || !raw.trim()) {
		return {
			type: "root",
			children: [],
		};
	}

	const markdown = texToMarkdown(raw);
	const ast = processor.parse(markdown);
	const hast = processor.runSync(ast) as FormattedContent;

	return hast;
}
