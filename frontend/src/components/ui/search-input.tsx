"use client";

import { Command, SearchIcon } from "lucide-react";
import React from "react";
import { Badge } from "src/components/ui/badge";

import { Input } from "src/components/ui/input";
import { useCommandPaletteContext } from "src/contexts/CommandPaletteContext";
import { cn } from "src/lib/utils";

interface SearchInputProps {
  placeholder?: string;
  className?: string;
}

export function SearchInput({
  placeholder = "Search pages, apps, etc...",
  className,
}: SearchInputProps) {
  const { setOpen } = useCommandPaletteContext();

  const handleClick = React.useCallback(() => {
    setOpen(true);
  }, [setOpen]);

  const handleFocus = React.useCallback(
    (e: React.FocusEvent<HTMLInputElement>) => {
      e.target.blur(); // Remove focus from input
      setOpen(true);
    },
    [setOpen]
  );

  const handleKeyDown = React.useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        setOpen(true);
      }
    },
    [setOpen]
  );

  return (
    <div
      className={cn("relative cursor-pointer", className)}
      onClick={handleClick}
    >
      <Input
        placeholder={placeholder}
        readOnly
        className="cursor-pointer pl-8 pr-8"
        onFocus={handleFocus}
        onKeyDown={handleKeyDown}
        tabIndex={0}
      />
      <SearchIcon className="absolute left-2 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
      <Badge
        variant="secondary"
        className="absolute right-2 top-1/2 transform -translate-y-1/2 "
      >
        <Command />K
      </Badge>
    </div>
  );
}
