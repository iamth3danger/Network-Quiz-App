// Client-side calls go to this same-origin proxy route, which forwards to the
// real backend (see src/app/api/backend/[...path]/route.ts). The backend URL
// lives server-side in API_BASE_URL, so it can change per deployment without a
// rebuild — nothing here is baked into the browser bundle.
const API_PROXY_PREFIX = "/api/backend";

// Join the proxy prefix with a backend path, e.g. apiUrl("/terms").
export function apiUrl(path: string): string {
  return `${API_PROXY_PREFIX}${path}`;
}
