import type { ReactNode } from "react";

/**
 * App Router templates are remounted by Next.js when a workspace route changes.
 * Keeping the transition here scopes it to page content; the persistent
 * workspace layout (including the sidebar/backside menu) stays untouched.
 */
export default function WorkspaceTemplate({ children }: { children: ReactNode }) {
  return <div className="workspace-page-transition">{children}</div>;
}
