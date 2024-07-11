import { SVGAttributes } from "react";

export function AppleIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      {...props}
      width="16"
      height="16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <g clipPath="url(#a)" stroke="currentColor" strokeLinejoin="round">
        <path d="M9.978.758c-.578.04-1.254.41-1.648.892-.359.437-.654 1.086-.54 1.717.632.02 1.285-.36 1.663-.85.354-.455.621-1.1.525-1.76ZM13.573 5.697c-.717-.898-1.724-1.42-2.675-1.42-1.256 0-1.787.601-2.66.601-.899 0-1.582-.599-2.668-.599-1.067 0-2.202.652-2.923 1.767-1.012 1.57-.839 4.52.802 7.035.587.9 1.371 1.91 2.397 1.92.912.008 1.17-.586 2.406-.592 1.236-.007 1.47.6 2.381.59 1.027-.008 1.854-1.13 2.44-2.029.422-.645.578-.97.905-1.697-2.374-.904-2.755-4.28-.405-5.576Z" />
      </g>
      <defs>
        <clipPath id="a">
          <path fill="currentColor" d="M0 0h16v16H0z" />
        </clipPath>
      </defs>
    </svg>
  );
}
