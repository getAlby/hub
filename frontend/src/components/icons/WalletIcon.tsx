import { SVGAttributes } from "react";

export function WalletIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M12 8a2 2 0 012-2h4a2 2 0 012 2v8a2 2 0 01-2 2h-4a2 2 0 01-2-2V8zm2-1a1 1 0 00-1 1v8a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1h-4z"
      />
      <path
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M5.5 6A1.5 1.5 0 004 7.5v9A1.5 1.5 0 005.5 18h10a1.5 1.5 0 001.5-1.5v-9A1.5 1.5 0 0015.5 6h-10zm2 7.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3z"
      />
    </svg>
  );
}
