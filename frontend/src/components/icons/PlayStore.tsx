import { SVGAttributes } from "react";

export function PlayStoreIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      {...props}
      width="16"
      height="16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 18 20"
    >
      <path
        d="M16.9 8.09L3.354.29a2.173 2.173 0 00-2.178.007A2.208 2.208 0 00.088 2.198v15.603c0 .78.416 1.51 1.087 1.901a2.171 2.171 0 002.178.008L16.9 11.908a2.206 2.206 0 000-3.817zm-5.87-1.844l-1.86 2.61-4.45-6.243 6.31 3.633zM2.185 18.658a.847.847 0 01-.346-.109.87.87 0 01-.428-.748V2.198a.856.856 0 01.772-.86L8.355 10l-6.17 8.658zm2.536-1.272l4.449-6.243 1.86 2.61-6.31 3.633zm11.523-6.635l-4.059 2.337L9.985 10l2.2-3.088 4.059 2.337a.868.868 0 010 1.502z"
        fill="currentColor"
      />
    </svg>
  );
}
