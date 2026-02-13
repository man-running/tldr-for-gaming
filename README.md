<img src="https://takara.ai/images/logo-24/TakaraAi.svg" width="200" alt="Takara.ai Logo" />

From the Frontier Research Team at **Takara.ai** we present **Takara TLDR**, a modern RSS-powered application that delivers the latest AI research papers in a beautifully designed interface. This application fetches and displays AI research summaries daily at 7am UTC from our papers API.

## Features

- **Daily Updates**: Fresh AI research summaries delivered daily at 7am UTC
- **Research Paper API**: High-performance API for fetching paper data from ArXiv and HuggingFace
- **Responsive Design**: Beautiful modern UI that works on all device sizes
- **Dark Mode Support**: Automatically adapts to your system preference
- **Accessibility**: Built with a11y best practices for all users

## Tech Stack

This Next.js application utilizes:

- **Next.js 15**: For server-side rendering and optimal performance
- **React 19**: For building the user interface components
- **Tailwind CSS**: For styling with a utility-first approach
- **TypeScript**: For type safety throughout the codebase
- **RSS Parser**: Custom implementation to process the research summaries feed
- **OpenAPI 3.0.3**: For comprehensive API documentation and type safety

## Local Development

```bash
# Install dependencies
bun install

# Start the development server
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## Debug Mode

To enable detailed RSS feed logging for development, set the environment variable:

```bash
RSS_DEBUG=true bun dev
```

This will show detailed logs about RSS feed fetching, parsing, and content processing. Only works in development mode.

## Deployment

The application is designed to be deployed on Vercel or similar platforms:

```bash
bun run build
```

## API Documentation

The application includes a comprehensive Paper API for fetching research paper data:

- **API Documentation**: See [API.md](./API.md) for complete documentation
- **OpenAPI Spec**: Available at `/api/openapi`
- **TypeScript Client**: Available in `types/api-client.ts`

### Quick API Example

```typescript
import { createPaperApiClient } from '@/types/api-client';

const client = createPaperApiClient();
const paper = await client.getPaper('1706.03762');
console.log(paper.title); // "Attention Is All You Need"
```

## Customization

Configuration values are stored in `lib/constants.ts` for easy customization of site details and feed sources.

## Dependencies

The project uses the following key dependencies:

- **fast-xml-parser**: For efficient RSS feed parsing
- **next-themes**: For dark/light mode theme switching
- **radix-ui**: For accessible UI components
- **lucide-react**: For beautiful icons

---

For research inquiries and press, please reach out to research@takara.ai

> 人類を変革する
