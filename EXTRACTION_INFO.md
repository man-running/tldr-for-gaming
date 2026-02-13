# TLDR Gaming - Standalone Application

This is a standalone extraction of the TLDR application from the takara-monorepo.

## Extraction Details

**Date:** February 13, 2026
**Source:** `/Volumes/Extreme/dev/takara-monorepo/apps/tldr`
**Destination:** `/Volumes/Extreme/dev/tldr-gaming`

## What Was Extracted

- **Main Application:** Complete Next.js application with all components, pages, and API routes
- **Shared UI Package:** Local copy of `@takara/shared-ui` dependency in `./packages/shared-ui`
- **Configuration Files:** All Next.js, TypeScript, and build configuration files
- **Assets & Public Files:** All public assets and static files

## Changes Made

### Package Configuration
- Updated `package.json` name: `takara-tldr` → `tldr-gaming`
- Changed dependency reference: `@takara/shared-ui: workspace:*` → `@takara/shared-ui: file:./packages/shared-ui`

### Path Updates
- **TypeScript Config** (`tsconfig.json`):
  - `@takara/shared-ui`: `../../packages/shared-ui/src` → `./packages/shared-ui/src`
  - `@takara/shared-ui/*`: `../../packages/shared-ui/src/*` → `./packages/shared-ui/src/*`

- **CSS Imports** (`app/globals.css`):
  - `../../../packages/shared-ui/src` → `../packages/shared-ui/src`

## Structure

```
tldr-gaming/
├── app/                          # Next.js app directory
├── api/                          # API routes
├── components/                   # React components
├── lib/                          # Shared libraries (including Go backend code)
├── packages/
│   └── shared-ui/               # Local UI component library
├── public/                       # Static assets
├── scripts/                      # Utility scripts
├── types/                        # TypeScript type definitions
├── package.json                  # Updated for standalone use
├── tsconfig.json                 # Updated paths
├── next.config.mjs              # Next.js configuration
└── ... (other config files)
```

## Getting Started

### Install Dependencies
```bash
bun install
```

### Development
```bash
bun dev
```

### Build
```bash
bun build
```

### Production
```bash
bun start
```

## Dependencies

The application uses:
- **Next.js 15.5.9** - React framework
- **React 19.2.1** - UI library
- **TypeScript 5.9.3** - Type safety
- **Tailwind CSS 4.1.18** - Styling
- **Bun 1.3.4** - Package manager and runtime

## Notes

- The application still maintains references to the original domain/branding (e.g., `tldr.takara.ai`)
- Go modules (`go.mod`, `go.sum`) are included for backend functionality
- The `.github` directory was copied (contains GitHub-specific workflows)
- `.cursorrules` file is included for AI-assisted development

## Next Steps

You may want to:
1. Update branding/domain references if needed
2. Initialize a new Git repository if desired
3. Update README.md with project-specific information
4. Adjust environment variables as needed
5. Configure deployment settings in `vercel.json` if needed
