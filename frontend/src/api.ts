// This file exists only to avoid path-resolution conflicts between `./api.ts` and `./api/`.
// Cloudflare/TypeScript resolves `import ... from "./api"` to `api.ts` first.
// We re-export everything from the real API module in `./api/index.ts`.

export * from "./api/index";
