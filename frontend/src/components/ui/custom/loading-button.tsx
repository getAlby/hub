import { Loader2Icon } from "lucide-react";
import * as React from "react";
import { Button } from "src/components/ui/button";

export interface ButtonProps extends React.ComponentProps<typeof Button> {
  loading?: boolean;
}

const LoadingButton = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant,
      size,
      asChild = false,
      loading,
      children,
      disabled,
      ...props
    },
    ref
  ) => {
    return (
      <Button
        className={className}
        variant={variant}
        size={size}
        asChild={asChild}
        disabled={disabled || loading}
        ref={ref}
        {...props}
      >
        {loading && <Loader2Icon className="size-4 animate-spin" />}
        {children}
      </Button>
    );
  }
);
LoadingButton.displayName = "LoadingButton";

export { LoadingButton };
