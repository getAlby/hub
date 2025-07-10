import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

import { cn } from "src/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-hidden focus:ring-2 focus:ring-ring focus:ring-offset-2",
  {
    variants: {
      variant: {
        default:
          "border-primary-foreground bg-primary text-primary-foreground shadow-sm hover:bg-primary/80",
        secondary:
          "border-secondary-foreground bg-secondary text-secondary-foreground hover:bg-secondary/80",
        destructive:
          "border-destructive-foreground bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/80",
        positive:
          "border-positive-foreground bg-positive text-positive-foreground",
        warning: "border-warning-foreground bg-warning text-warning-foreground",
        outline: "text-foreground",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export { Badge, badgeVariants };
