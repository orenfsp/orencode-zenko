import type { NextConfig } from "next";

const backendUrl = process.env.BACKEND_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      // Keep the rank collection URL exact: the backend registers it without
      // a trailing slash and Go's redirect can otherwise bounce through the
      // frontend proxy.
      {
        source: "/api/v1/rank",
        destination: `${backendUrl}/api/v1/rank`,
      },
      {
        source: "/api/v1/:path*",
        destination: `${backendUrl}/api/v1/:path*`,
      },
    ];
  },
};

export default nextConfig;
