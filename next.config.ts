import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Emit a self-contained build at .next/standalone for a minimal Docker image.
  output: "standalone",
  // Pin the trace root to this app. Otherwise Next walks up to the
  // network_study parent (Go backend + .git) and nests the output under
  // .next/standalone/network_study/quiz-app, breaking the Dockerfile COPY paths.
  outputFileTracingRoot: __dirname,
};

export default nextConfig;
