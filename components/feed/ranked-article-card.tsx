"use client";

import type { ReactNode } from "react";

interface RankedArticleCardProps {
  rank: number;
  title: string;
  summary: string;
  originalSummary: string;
  sourceName: string;
  url: string;
  score: number;
  reason: string;
  publishedDate: string;
  categories: string[];
  imageUrl?: string;
}

export function RankedArticleCard({
  rank,
  title,
  summary,
  originalSummary,
  sourceName,
  url,
  score,
  reason,
  publishedDate,
  categories,
  imageUrl,
}: RankedArticleCardProps) {
  // Format the published date
  const publishedDateObj = new Date(publishedDate);
  const formattedDate = publishedDateObj.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });

  // Get badge color based on rank
  const rankColors = [
    "bg-yellow-100 text-yellow-800 border-yellow-300", // 1st
    "bg-gray-100 text-gray-800 border-gray-300", // 2nd
    "bg-orange-100 text-orange-800 border-orange-300", // 3rd
    "bg-blue-100 text-blue-800 border-blue-300", // 4th
    "bg-purple-100 text-purple-800 border-purple-300", // 5th
  ];

  const rankColor = rankColors[rank - 1] || rankColors[4];

  // Get reason badge colors
  const reasonBadgeColors: Record<string, string> = {
    trending: "bg-red-100 text-red-800",
    authoritative: "bg-green-100 text-green-800",
    "high-engagement": "bg-blue-100 text-blue-800",
    diverse: "bg-purple-100 text-purple-800",
    featured: "bg-gray-100 text-gray-800",
  };

  const getReasonColor = (r: string): string => {
    const normalized = r.toLowerCase();
    if (normalized.includes("trending")) return reasonBadgeColors.trending;
    if (normalized.includes("authoritative"))
      return reasonBadgeColors.authoritative;
    if (normalized.includes("engagement"))
      return reasonBadgeColors["high-engagement"];
    if (normalized.includes("diverse")) return reasonBadgeColors.diverse;
    return reasonBadgeColors.featured;
  };

  return (
    <article className="flex flex-col gap-4 rounded-lg border border-gray-200 bg-white p-6 shadow-sm transition-shadow hover:shadow-md dark:border-gray-700 dark:bg-gray-900">
      {/* Header with Rank Badge and Score */}
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <div className="flex items-center gap-3 mb-3">
            <span
              className={`inline-flex h-8 w-8 items-center justify-center rounded-full font-bold border ${rankColor}`}
            >
              #{rank}
            </span>
            <span className="text-sm font-medium text-gray-600 dark:text-gray-400">
              Score: {(score * 100).toFixed(0)}%
            </span>
          </div>
          <h3 className="text-lg font-bold leading-tight text-gray-900 dark:text-white">
            <a
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-blue-600 dark:hover:text-blue-400"
            >
              {title}
            </a>
          </h3>
        </div>

        {/* Image if available */}
        {imageUrl && (
          <img
            src={imageUrl}
            alt={title}
            className="h-24 w-32 rounded object-cover"
          />
        )}
      </div>

      {/* Reason Badge */}
      <div>
        <span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${getReasonColor(reason)}`}>
          {reason.split(",")[0]}
        </span>
      </div>

      {/* AI Summary (if available) */}
      {summary && (
        <div className="bg-blue-50 dark:bg-blue-900/20 rounded-md border border-blue-200 dark:border-blue-800 p-4">
          <h4 className="text-sm font-semibold text-blue-900 dark:text-blue-200 mb-2">
            AI Summary
          </h4>
          <p className="text-sm text-blue-800 dark:text-blue-300">{summary}</p>
        </div>
      )}

      {/* Original Summary */}
      {originalSummary && !summary && (
        <p className="text-sm text-gray-700 dark:text-gray-300">
          {originalSummary}
        </p>
      )}

      {/* Categories and Source */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-gray-200 pt-4 dark:border-gray-700">
        <div className="flex flex-wrap gap-2">
          {categories.map((cat) => (
            <span
              key={cat}
              className="inline-block rounded-full bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-700 dark:bg-gray-800 dark:text-gray-300"
            >
              {cat}
            </span>
          ))}
        </div>

        <div className="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
          <span className="font-semibold text-gray-700 dark:text-gray-300">
            {sourceName}
          </span>
          <span>{formattedDate}</span>
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            className="font-semibold text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
          >
            Read →
          </a>
        </div>
      </div>
    </article>
  );
}
