import { SVGAttributes } from "react";

export function PlayStoreIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      {...props}
      width="16"
      height="16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M2.835 15.388a.894.894 0 0 1-.252.094l7.048-6.787 1.973 1.897-8.766 4.795-.003.001Zm9.689-5.299-2.172-2.088 2.172-2.087 2.359 1.291h.001a.89.89 0 0 1 .465.791v.01a.89.89 0 0 1-.465.791h-.001l-2.359 1.292Zm-.92-4.678L9.631 7.308 2.58.518c.086.018.171.05.251.095l.006.003 8.766 4.795ZM8.91 8l-7.3 7.03a.866.866 0 0 1-.109-.429V1.401a.867.867 0 0 1 .11-.43l7.3 7.03Z"
        stroke="currentColor"
      />
    </svg>
  );
}
