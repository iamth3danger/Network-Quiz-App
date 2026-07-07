# syntax=docker/dockerfile:1

FROM node:26-slim AS base

# Install dependencies only when needed
FROM base AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# Build the app
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# Production runner
FROM base AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Run as a non-root user
RUN groupadd --system --gid 1001 nodejs \
  && useradd --system --uid 1001 --gid nodejs nextjs

# The standalone output already bundles a minimal node_modules and server.js.
# public and .next/static must be copied alongside it (server.js serves them).
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000
ENV PORT=3000
# Bind to all interfaces so the container is reachable from outside.
ENV HOSTNAME=0.0.0.0
# Backend URL is read at runtime; override at `docker run -e API_BASE_URL=...`.
ENV API_BASE_URL=http://localhost:8080

CMD ["node", "server.js"]
