import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";
import { Link } from "react-router-dom";

import ExternalLink from "src/components/ExternalLink";
import { cn } from "src/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground shadow hover:bg-primary/90",
        destructive:
          "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90",
        destructive_outline:
          "border border-destructive text-destructive shadow-sm hover:text-destructive/90",
        outline:
          "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground",
        secondary:
          "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-foreground hover:text-accent-foreground underline-offset-4 hover:underline",
        positive:
          "bg-positive text-positive-foreground shadow-sm hover:bg-positive/90",
        premium:
          "text-black shadow-md shadow-amber-500/20 bg-gradient-to-r from-amber-500 to-amber-300 hover:from-amber-400 hover:to-amber-200 transition-all duration-200 hover:shadow-lg hover:shadow-amber-500/30",
      },
      size: {
        default: "h-9 px-4 py-2",
        sm: "h-8 rounded-md px-3 text-xs",
        lg: "h-10 rounded-md px-8",
        icon: "h-9 w-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";

type LinkProps = React.PropsWithChildren<
  VariantProps<typeof buttonVariants> & {
    to: string;
    className?: string;
  }
>;
const ExternalLinkButton = React.forwardRef<HTMLAnchorElement, LinkProps>(
  ({ className, variant, size, ...props }, _ref) => {
    return (
      <ExternalLink
        className={cn(buttonVariants({ variant, size, className }))}
        {...props}
      />
    );
  }
);

const LinkButton = React.forwardRef<HTMLAnchorElement, LinkProps>(
  ({ className, variant, size, ...props }, _ref) => {
    return (
      <Link
        className={cn(buttonVariants({ variant, size, className }))}
        {...props}
      />
    );
  }
);

//const ExternalLinkButton = LinkButton
// eslint-disable-next-line react-refresh/only-export-components
export { Button, buttonVariants, ExternalLinkButton, LinkButton };
