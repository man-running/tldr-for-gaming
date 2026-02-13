# @takara/shared-ui

Shared UI components and utilities for Takara monorepo applications.

## Purpose

This package provides shared UI components and utilities that can be used across both the `website` and `tldr` applications to maintain design consistency.

## Usage

### Importing Components

```typescript
// Import Button component
import { Button } from "@takara/shared-ui/button";

// Import Badge component
import { Badge } from "@takara/shared-ui/badge";

// Import utilities
import { cn } from "@takara/shared-ui/utils";

// Or import everything from the main entry
import { Button, Badge, cn } from "@takara/shared-ui";
```

### Example

```typescript
import { Button } from "@takara/shared-ui/button";

export function MyComponent() {
  return (
    <Button variant="default" size="lg">
      Click me
    </Button>
  );
}
```

## Available Components

- **Button**: Fully styled button component with multiple variants and sizes
- **Badge**: Badge component for labels and tags

## Available Utilities

- **cn**: Utility function for merging Tailwind CSS classes (using `clsx` and `tailwind-merge`)

## Adding New Components

1. Create the component file in `src/components/`
2. Export it from `src/index.ts`
3. Add it to the package.json exports field
4. Update this README

## Development

This package is part of the Bun workspace monorepo. Dependencies are managed at the root level.

To add this package as a dependency in an app:

```json
{
  "dependencies": {
    "@takara/shared-ui": "workspace:*"
  }
}
```

Then configure Next.js to transpile the package:

```typescript
// next.config.ts
const nextConfig = {
  transpilePackages: ["@takara/shared-ui"],
  // ... other config
};
```

And add TypeScript path mapping:

```json
// tsconfig.json
{
  "compilerOptions": {
    "paths": {
      "@takara/shared-ui": ["../../packages/shared-ui/src"],
      "@takara/shared-ui/*": ["../../packages/shared-ui/src/*"]
    }
  }
}
```

