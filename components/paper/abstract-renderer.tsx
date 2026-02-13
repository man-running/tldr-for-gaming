import type { Element, Node, Root, Text } from "hast";
import type { ReactNode } from "react";

interface AbstractRendererProps {
	content: Root;
}

const SENTENCES_PER_BREAK = 2;

function insertLineBreaksAfterSentences(
	text: string,
	index: number = 0,
): ReactNode {
	if (!text) return text;

	const parts: ReactNode[] = [];
	const sentences = text.split(/(\.\s+)/);

	if (sentences.length <= 1) {
		return text;
	}

	let sentenceCount = 0;
	let currentGroup: string[] = [];
	let partIndex = index;

	for (let i = 0; i < sentences.length; i++) {
		const segment = sentences[i];
		currentGroup.push(segment);

		if (segment.match(/\.\s+/)) {
			sentenceCount++;

			if (sentenceCount >= SENTENCES_PER_BREAK) {
				const groupText = currentGroup.join("");
				parts.push(<span key={`text-${partIndex++}`}>{groupText}</span>);
				parts.push(<br key={`br-${partIndex++}`} />);
				parts.push(<br key={`br-${partIndex++}`} />);
				currentGroup = [];
				sentenceCount = 0;
			}
		}
	}

	if (currentGroup.length > 0) {
		parts.push(
			<span key={`text-${partIndex++}`}>{currentGroup.join("")}</span>,
		);
	}

	return parts.length > 0 ? parts : text;
}

function renderNode(node: Node | undefined, index: number = 0): ReactNode {
	if (!node) return null;

	if (node.type === "root") {
		const rootNode = node as Root;
		return (
			<>
				{rootNode.children?.map((child, i) => (
					<div key={`${child.type}-${index}-${i}`}>
						{renderNode(child, index + i)}
					</div>
				))}
			</>
		);
	}

	if (node.type === "element") {
		const element = node as Element;
		const children = element.children?.flatMap((child, i) => {
			const result = renderNode(child, index + i);
			if (Array.isArray(result)) {
				return result;
			}
			return result != null ? [result] : [];
		});

		switch (element.tagName) {
			case "p":
				return <p>{children}</p>;
			case "a":
				return (
					<a
						href={element.properties?.href as string | undefined}
						target="_blank"
						rel="noopener noreferrer"
					>
						{children}
					</a>
				);
			case "strong":
				return <strong>{children}</strong>;
			case "em":
				return <em>{children}</em>;
			case "code":
				return <code>{children}</code>;
			default:
				return <>{children}</>;
		}
	}

	if (node.type === "text") {
		const textNode = node as Text;
		return insertLineBreaksAfterSentences(textNode.value, index);
	}

	return null;
}

export function AbstractRenderer({ content }: AbstractRendererProps) {
	if (!content || content.children.length === 0) {
		return null;
	}

	return (
		<div className="prose mx-auto w-full">
			<h2 className="text-headline-md text-foreground mb-4">Abstract</h2>
			{renderNode(content)}
		</div>
	);
}
