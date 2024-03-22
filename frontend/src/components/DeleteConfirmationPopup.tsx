import { PopiconsTriangleExclamationLine } from "@popicons/react";

type Props = {
  appName: string;
  onConfirm: () => void;
  onCancel: () => void;
};

function DeleteConfirmationPopup({ appName, onConfirm, onCancel }: Props) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-10 flex justify-center items-center">
      <div className="rounded-xl mx-2 w-full max-w-lg bg-white border flex flex-col justify-between">
        <div className="p-4 h-full border-b flex items-center">
          <PopiconsTriangleExclamationLine className="w-20 h-20 text-red-300" />
          <div className="ml-4">
            <h2 className="font-medium text-gray-800 mb-2">
              Disconnecting <span className="font-bold">{appName}</span>
            </h2>
            <p className="text-sm text-gray-500">
              This will revoke the permission and will no longer allow calls
              from this public key.
            </p>
          </div>
        </div>
        <div className="py-3 px-4 flex items-center gap-4">
          <button
            onClick={onCancel}
            className="text-center font-medium p-2.5 w-full text-sm rounded-lg text-gray-700 hover:bg-gray-100 cursor-pointer whitespace-nowrap"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="text-center font-medium p-2.5 w-full text-sm rounded-lg text-white bg-red-500 cursor-pointer hover:bg-red-600 whitespace-nowrap"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  );
}

export default DeleteConfirmationPopup;
