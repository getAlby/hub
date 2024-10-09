import { SVGAttributes } from "react";

export function FirefoxIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      fill="none"
      {...props}
    >
      <path
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M2.633 3.52a6.994 6.994 0 0 1 8.162-1.923"
      />
      <path
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M15 8.013a6.998 6.998 0 0 1-11.582 5.328 7 7 0 0 1-1.505-8.749 1.88 1.88 0 0 1 .304-2.361c.05.453.191.892.414 1.29.177.329.365.566.5.536a5.35 5.35 0 0 1 1.521 0c.25-.323.858-1.218 1.826-1.218-.542.542-2.094 2.435.609 2.435h.609L5.869 6.491s.207.402 0 .609c-.23-.231-1.217.25-1.217.913s.706 1.522 2.13 1.522c1.425 0 1.066-.609 1.827-.609a.883.883 0 0 1 .913.609c-1.047 0-1.826 1.217-3.044 1.217a2.6 2.6 0 0 0 1.826.609 3.653 3.653 0 0 0 2.934-5.825 3.653 3.653 0 0 1 1.461 2.61 5.44 5.44 0 0 0 .17-1.35c0-1.826-.73-4.261-2.075-5.198A6.988 6.988 0 0 1 15 8.013Z"
      />
    </svg>
  );
}
