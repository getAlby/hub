import * as React from "react";
import { type VariantProps } from "class-variance-authority";
import { Slot } from "radix-ui";

import { cn } from "src/lib/utils";
import { badgeVariants } from "src/components/ui/badgeVariants";

function Badge({
  className,
  variant = "default",
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot.Root : "span";

  return (
    <Comp
      data-slot="badge"
      data-variant={variant}
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  );
}

export { Badge };
