import { wordlist } from "@scure/bip39/wordlists/english.js";
import { useState } from "react";
import RevealPasswordToggle from "src/components/password/RevealPasswordToggle";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { InputWithAdornment } from "src/components/ui/custom/input-with-adornment";

type MnemonicInputsProps = {
  mnemonic?: string;
  setMnemonic?(mnemonic: string): void;
  readOnly?: boolean;
  asCard?: boolean;
};

export default function MnemonicInputs({
  mnemonic,
  setMnemonic,
  readOnly,
  children,
  asCard = true,
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

  const content = (
    <>
      <div className="grid grid-cols-2 gap-y-5 gap-8 justify-center backup sensitive">
        {words.map((word, i) => {
          const isRevealed = revealedIndex === i;
          const inputId = `mnemonic-word-${i}`;
          return (
            <div key={i} className="flex justify-center items-center gap-2">
              <span className="text-foreground text-right">{i + 1}.</span>
              <div className="relative">
                <InputWithAdornment
                  id={inputId}
                  autoFocus={!readOnly && i === 0}
                  readOnly={readOnly}
                  onFocus={() => setRevealedIndex(i)}
                  onBlur={() => setRevealedIndex(undefined)}
                  className="w-32 sm:w-44 border-[#E4E4E7] text-muted-foreground"
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
                    <RevealPasswordToggle
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
    </>
  );
  if (!asCard) {
    return content;
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="text-xl text-center">
            Wallet Recovery Phrase
          </CardTitle>
        </CardHeader>

        <CardContent>{content}</CardContent>
      </Card>
    </>
  );
}
