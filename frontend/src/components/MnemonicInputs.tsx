import { wordlist } from "@scure/bip39/wordlists/english";
import { useState } from "react";
import PasswordViewAdornment from "src/components/password/PasswordAdornment";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";

type MnemonicInputsProps = {
  mnemonic?: string;
  setMnemonic?(mnemonic: string): void;
  readOnly?: boolean;
  description?: string;
};

export default function MnemonicInputs({
  mnemonic,
  setMnemonic,
  readOnly,
  children,
  description,
}: React.PropsWithChildren<MnemonicInputsProps>) {
  const words = mnemonic?.split(" ") || [];
  while (words.length < 12) {
    words.push("");
  }

  while (words.length > 12) {
    words.pop();
  }

  const [revealedIndex, setRevealedIndex] = useState<number | undefined>(
    undefined
  );

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="text-xl text-center">
            Wallet Recovery Phrase
          </CardTitle>
          <CardDescription className="text-muted-foreground text-sm max-w-xs text-right mt-2">
            {description}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-y-5 justify-center backup sensitive">
            {words.map((word, i) => {
              const isRevealed = revealedIndex === i;
              const inputId = `mnemonic-word-${i}`;
              return (
                <div key={i} className="flex justify-center items-center gap-2">
                  <span className="text-foreground text-right">{i + 1}.</span>
                  <div className="relative">
                    <Input
                      id={inputId}
                      autoFocus={!readOnly && i === 0}
                      readOnly={readOnly}
                      onFocus={() => setRevealedIndex(i)}
                      onBlur={() => setRevealedIndex(undefined)}
                      className="w-32 text-center border-[#E4E4E7] text-muted-foreground"
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
                      endAdornment={
                        <PasswordViewAdornment
                          isRevealed={isRevealed}
                          onChange={(passwordView) => {
                            if (passwordView) {
                              document.getElementById(inputId)?.focus();
                            }
                          }}
                          iconClass="text-muted-foreground"
                        />
                      }
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
