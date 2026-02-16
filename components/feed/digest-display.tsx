"use client";

import { useEffect, useState } from "react";

interface Article {
  id: string;
  title: string;
  summary: string;
  originalSummary: string;
  url: string;
  sourceName: string;
  sourceId: string;
  publishedDate: string;
  imageUrl?: string;
  categories?: string[];
}

interface RankedArticle {
  article: Article;
  score: number;
  rank: number;
  reason: string;
}

interface DailyDigest {
  date: string;
  articles: RankedArticle[];
  headline: string;
  summary: string;
  created: string;
}

interface DigestDisplayProps {
  date?: string;
}

// Convert markdown links to HTML links
function markdownToHtml(text: string): React.ReactNode {
  const linkRegex = /\[([^\]]+)\]\(([^)]+)\)/g;
  const parts: React.ReactNode[] = [];
  let lastIndex = 0;

  let match;
  while ((match = linkRegex.exec(text)) !== null) {
    // Add text before link
    if (match.index > lastIndex) {
      parts.push(text.slice(lastIndex, match.index));
    }

    // Add link
    const [fullMatch, title, url] = match;
    parts.push(
      <a
        key={`link-${match.index}`}
        href={url}
        className="font-bold underline hover:text-blue-600 dark:hover:text-blue-400"
      >
        {title}
      </a>
    );

    lastIndex = match.index + fullMatch.length;
  }

  // Add remaining text
  if (lastIndex < text.length) {
    parts.push(text.slice(lastIndex));
  }

  return parts;
}

export function DigestDisplay({ date }: DigestDisplayProps) {
  const [digest, setDigest] = useState<DailyDigest | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchDigest = async () => {
      try {
        setLoading(true);
        setError(null);

        const url = date
          ? `/api/digest?date=${date}`
          : "/api/digest";

        const response = await fetch(url);
        if (!response.ok) {
          throw new Error(`Failed to fetch digest: ${response.statusText}`);
        }

        const data: DailyDigest = await response.json();
        setDigest(data);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load digest"
        );
      } finally {
        setLoading(false);
      }
    };

    fetchDigest();
  }, [date]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-600 dark:text-gray-400">
          Loading...
        </p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="py-8">
        <p className="text-red-600 dark:text-red-400">{error}</p>
      </div>
    );
  }

  if (!digest) {
    return (
      <div className="py-8">
        <p className="text-gray-600 dark:text-gray-400">
          No digest available.
        </p>
      </div>
    );
  }

  // Format date for display (e.g., "February 16, 2026")
  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  return (
    <div className="max-w-3xl mx-auto">
      {/* Banner with Title, Date, and Headline */}
      <div className="mb-12 border-b border-gray-200 dark:border-gray-700 pb-8">
        <h1 className="text-5xl md:text-6xl font-bold text-gray-900 dark:text-white mb-3">
          TLDR: Gaming and Gambling
        </h1>
        <p className="text-lg text-gray-600 dark:text-gray-400 mb-4">
          {formatDate(digest.date)}
        </p>
        <p className="text-xl md:text-2xl font-semibold text-gray-800 dark:text-gray-200 leading-relaxed">
          {digest.headline}
        </p>
      </div>

      {/* Narrative Summary */}
      <div className="space-y-4 text-gray-700 dark:text-gray-300 leading-relaxed">
        {digest.summary.split("\n\n").map((paragraph, idx) => (
          <p key={idx} className="text-gray-700 dark:text-gray-300">
            {markdownToHtml(paragraph)}
          </p>
        ))}
      </div>
    </div>
  );
}
