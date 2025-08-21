import type { VariantProps } from "class-variance-authority";
import * as React from "react";
import ExternalLink from "src/components/ExternalLink";
import { cn } from "src/lib/utils";
import { buttonVariants } from "../buttonVariants";

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
