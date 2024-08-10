import { wordlist } from "@scure/bip39/wordlists/english";
import { useState } from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";

type MnemonicInputsProps = {
  mnemonic?: string;
  setMnemonic?(mnemonic: string): void;
  readOnly?: boolean;
};

export default function MnemonicInputs({
  mnemonic,
  setMnemonic,
  readOnly,
  children,
}: React.PropsWithChildren<MnemonicInputsProps>) {
  const [revealedIndex, setRevealedIndex] = useState<number | undefined>(
    undefined
  );

  const words = mnemonic?.split(" ") || [];
  while (words.length < 12) {
    words.push("");
  }

  while (words.length > 12) {
    words.pop();
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Recovery phrase to your wallet</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-x-8 gap-y-5 justify-center backup sensitive">
            {words.map((word, i) => {
              const isRevealed = revealedIndex === i;
              const inputId = `mnemonic-word-${i}`;
              return (
                <div key={i} className="flex justify-center items-center gap-2">
                  <span className="text-muted-foreground text-right">
                    {i + 1}.
                  </span>
                  <div className="relative">
                    <Input
                      id={inputId}
                      autoFocus={!readOnly && i === 0}
                      onFocus={() => setRevealedIndex(i)}
                      onBlur={() => setRevealedIndex(undefined)}
                      readOnly={readOnly}
                      className="w-32 text-center"
                      list={readOnly ? undefined : "wordlist"}
                      value={isRevealed ? word : word.length ? "•••••" : ""}
                      onChange={(e) => {
                        if (revealedIndex !== i) {
                          return;
                        }
                        words[i] = e.target.value;
                        setMnemonic?.(
                          words
                            .map((word) => word.trim())
                            .join(" ")
                            .trim()
                        );
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
          {!readOnly && (
            <datalist id="wordlist">
              {wordlist.map((word) => (
                <option key={word} value={word} />
              ))}
            </datalist>
          )}
          {children}
        </CardContent>
      </Card>
    </>
  );
}
