function Loading({ color }: { color?: string }) {
  return (
    <>
      <svg
        className="animate-spin h-6 w-6 text-purple-700"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          className="opacity-80"
          cx="12"
          cy="12"
          r="10"
          stroke={color ?? "#FFC700"}
          strokeWidth="4"
        ></circle>
        <path
          fill={color ?? "#897FFF"}
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        ></path>
      </svg>
      <span className="sr-only">Loading...</span>
    </>
  );
}

export default Loading;
