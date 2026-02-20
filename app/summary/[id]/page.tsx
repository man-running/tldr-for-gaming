'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import { Footer } from '@/components/layout/footer';

interface SummaryData {
  id: string;
  title: string;
  summary: string;
  url: string;
  sourceName: string;
  publishedDate: string;
  generatedAt: string;
  imageUrl?: string;
}

export default function SummaryPage({ params }: { params: Promise<{ id: string }> }) {
  const [summary, setSummary] = useState<SummaryData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [id, setId] = useState<string | null>(null);

  useEffect(() => {
    // First, resolve the params promise
    params.then(({ id: resolvedId }) => {
      setId(resolvedId);
    });
  }, [params]);

  useEffect(() => {
    if (!id) return;

    const fetchSummary = async () => {
      try {
        setLoading(true);
        const response = await fetch(`/api/summary/${id}`);

        if (!response.ok) {
          throw new Error('Summary not found');
        }

        const data = await response.json();
        setSummary(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load summary');
      } finally {
        setLoading(false);
      }
    };

    fetchSummary();
  }, [id]);

  if (loading) {
    return (
      <>
        <main className="min-h-screen w-full py-12">
          <div className="max-w-3xl mx-auto px-6">
            <p className="text-gray-600 dark:text-gray-400">Loading...</p>
          </div>
        </main>
        <Footer />
      </>
    );
  }

  if (error || !summary) {
    return (
      <>
        <main className="min-h-screen w-full py-12">
          <div className="max-w-3xl mx-auto px-6">
            <p className="text-red-600 dark:text-red-400">
              {error || 'Summary not found'}
            </p>
          </div>
        </main>
        <Footer />
      </>
    );
  }

  return (
    <>
      <main className="min-h-screen w-full">
        <div className="py-12 px-6 sm:px-8">
          <div className="max-w-3xl mx-auto">
            {/* Back arrow */}
            <a
              href="/gaming"
              className="inline-flex items-center gap-2 text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 mb-6 transition-colors"
            >
              <svg
                className="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 19l-7-7 7-7"
                />
              </svg>
              <span className="text-sm font-medium">Back</span>
            </a>

            {/* Banner Image or Gradient Placeholder */}
            <div className="relative w-full h-64 sm:h-80 md:h-96 mb-8 rounded-lg overflow-hidden">
              {summary.imageUrl ? (
                <Image
                  src={summary.imageUrl}
                  alt={summary.title}
                  fill
                  className="object-cover"
                  priority
                />
              ) : (
                <div className="absolute inset-0 bg-gradient-to-br from-accent to-pink-500">
                  <div className="absolute inset-0 flex items-center justify-center p-8">
                    <h2 className="text-2xl sm:text-3xl md:text-4xl font-bold text-white text-center drop-shadow-lg">
                      {summary.title}
                    </h2>
                  </div>
                </div>
              )}
            </div>

            {/* Title */}
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
              {summary.title}
            </h1>

            {/* Metadata */}
            <div className="text-sm text-gray-500 dark:text-gray-400 mb-8 space-y-1">
              <p>Source: {summary.sourceName}</p>
              <p>
                Published:{' '}
                {new Date(summary.publishedDate).toLocaleDateString('en-US', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </p>
            </div>

            {/* Summary */}
            <div className="prose prose-sm dark:prose-invert max-w-none mb-8">
              {summary.summary.split('\n').map((paragraph, idx) => (
                <p
                  key={idx}
                  className="text-gray-700 dark:text-gray-300 leading-relaxed mb-4"
                >
                  {paragraph}
                </p>
              ))}
            </div>

            {/* Divider */}
            <div className="border-t border-gray-200 dark:border-gray-700 my-8" />

            {/* Link to full article */}
            <div className="flex items-center gap-4">
              <a
                href={summary.url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 px-4 py-2 bg-accent hover:opacity-90 text-white font-medium rounded-lg transition-opacity"
              >
                Read Full Article
                <svg
                  className="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                  />
                </svg>
              </a>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
