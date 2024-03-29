import Loading from "src/components/Loading";

type ConnectButtonProps = {
  isConnecting: boolean;
  loadingText?: string;
  submitText?: string;
  disabled?: boolean;
};

export default function ConnectButton({
  isConnecting,
  loadingText,
  submitText,
  disabled,
}: ConnectButtonProps) {
  return (
    <button
      type="submit"
      className={`mt-4 inline-flex w-full gap-2 ${
        isConnecting || disabled
          ? "bg-gray-300 dark:bg-gray-700 pointer-events-none"
          : "bg-purple-700 hover:bg-purple-900"
      } cursor-pointer items-center justify-center rounded-md px-5 py-3 font-medium text-white shadow transition duration-150 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 dark:text-neutral-200`}
      disabled={isConnecting || disabled}
    >
      {isConnecting ? (
        <>
          <Loading className="text-white" /> {loadingText || "Connecting..."}
        </>
      ) : (
        <>{submitText || "Connect"}</>
      )}
    </button>
  );
}
