import type { VariantProps } from "class-variance-authority";
import * as React from "react";
import { Link } from "react-router-dom";
import { buttonVariants } from "src/components/ui/button";
import { cn } from "src/lib/utils";

export function LinkButton({
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
    <Link
      to={to}
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    >
      {children}
    </Link>
  );
}
