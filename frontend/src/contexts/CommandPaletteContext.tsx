import React from "react";
import { useCommandPalette } from "src/hooks/useCommandPalette";

interface CommandPaletteContextType {
  open: boolean;
  setOpen: (open: boolean) => void;
}

const CommandPaletteContext =
  React.createContext<CommandPaletteContextType | null>(null);

export function CommandPaletteProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const commandPalette = useCommandPalette();

  return (
    <CommandPaletteContext.Provider value={commandPalette}>
      {children}
    </CommandPaletteContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useCommandPaletteContext() {
  const context = React.useContext(CommandPaletteContext);
  if (!context) {
    throw new Error(
      "useCommandPaletteContext must be used within CommandPaletteProvider"
    );
  }
  return context;
}
