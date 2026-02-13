import type { FeedItemType } from "@/types/feed";
import { ContentRendererWithLinks } from "./content-renderer-with-links";

interface FeedItemProps {
	item: FeedItemType;
	index: number;
}

export function FeedItem({ item, index }: FeedItemProps) {
	const itemId = `item-${index}`;
	const isFirstSection = index === 0;

	return (
		<article id={itemId} className="w-full mb-8">
			<ContentRendererWithLinks
				html={item.description}
				isFirstSection={isFirstSection}
			/>
		</article>
	);
}
