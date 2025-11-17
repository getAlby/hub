import { VariantProps } from "class-variance-authority";
import * as React from "react";
import { cn } from "src/lib/utils";
import { Button } from "./ui/button";
import { buttonVariants } from "./ui/buttonVariants";

type Props = {
  icon: React.ComponentType;
  text: string;
} & React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  };

const ResponsiveButton = ({
  icon: Icon,
  text,
  variant,
  asChild,
  children,
  className,
  ...props
}: Props) => {
  const content = (
    <>
      <Icon />
      <span className="hidden lg:inline">{text}</span>
    </>
  );

  return (
    <Button
      {...props}
      variant={variant}
      asChild={asChild}
      className={cn(
        className,
        "max-lg:size-9" /* apply size="icon" only for mobile */
      )}
    >
      {asChild
        ? React.cloneElement(children as React.ReactElement, {}, content)
        : content}
    </Button>
  );
};

export default ResponsiveButton;
