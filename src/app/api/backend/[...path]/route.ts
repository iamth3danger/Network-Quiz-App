import { type NextRequest, NextResponse } from "next/server";

// Backend base URL, read on the server at request time (not inlined into the
// client bundle). Set API_BASE_URL in .env / the container environment; it can
// change per deployment without rebuilding.
const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080";

// Proxy every /api/backend/* request through to the Go backend. The browser
// only ever calls this same-origin route, so the real backend URL stays server-side.
async function proxy(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> },
) {
  const { path } = await params;
  const target = `${API_BASE_URL}/${path.join("/")}${request.nextUrl.search}`;

  const hasBody = request.method !== "GET" && request.method !== "HEAD";
  const init: RequestInit = {
    method: request.method,
    headers: request.headers.get("content-type")
      ? { "content-type": request.headers.get("content-type")! }
      : undefined,
    body: hasBody ? await request.text() : undefined,
  };

  try {
    const res = await fetch(target, init);
    const body = await res.text();
    return new NextResponse(body, {
      status: res.status,
      headers: {
        "content-type": res.headers.get("content-type") ?? "application/json",
      },
    });
  } catch (err) {
    console.error(`Proxy to ${target} failed:`, err);
    return NextResponse.json(
      { status: 502, error: "backend unreachable" },
      { status: 502 },
    );
  }
}

export {
  proxy as GET,
  proxy as POST,
  proxy as PUT,
  proxy as PATCH,
  proxy as DELETE,
};
