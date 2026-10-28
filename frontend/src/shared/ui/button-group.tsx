import * as React from "react";

import { cn } from "@/shared/lib/utils";

function ButtonGroup({ className, orientation = "horizontal", ...props }: React.ComponentProps<"div"> & { orientation?: "horizontal" | "vertical" }) {
  return <div role="group" data-slot="button-group" data-orientation={orientation} className={cn("flex w-fit items-stretch [&>*]:focus-visible:relative [&>*]:focus-visible:z-10", orientation === "vertical" ? "flex-col [&>[data-slot=button]:not(:first-child)]:rounded-t-none [&>[data-slot=button]:not(:last-child)]:rounded-b-none" : "flex-row [&>[data-slot=button]:not(:first-child)]:rounded-l-none [&>[data-slot=button]:not(:last-child)]:rounded-r-none", className)} {...props} />;
}

function ButtonGroupText({ className, ...props }: React.ComponentProps<"span">) {
  return <span data-slot="button-group-text" className={cn("px-2 text-sm text-muted-foreground", className)} {...props} />;
}

export { ButtonGroup, ButtonGroupText };
