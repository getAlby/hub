import type { VariantProps } from "class-variance-authority";
import * as React from "react";
import ExternalLink from "src/components/ExternalLink";
import { buttonVariants } from "src/components/ui/button";
import { cn } from "src/lib/utils";

export function ExternalLinkButton({
  className,
  variant,
  size,
  to,
  children,
  ...props
}: React.PropsWithChildren<
  VariantProps<typeof buttonVariants> & {
    to: string;
    className?: string;
  }
>) {
  return (
    <ExternalLink
      to={to}
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    >
      {children}
    </ExternalLink>
  );
}
