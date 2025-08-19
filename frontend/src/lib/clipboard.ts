import { toast } from "sonner";

// TODO: change this function to a hook and then toast does not need to be passed in
export async function copyToClipboard(content: string) {
  const copyPromise = new Promise((resolve, reject) => {
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(content).then(resolve).catch(reject);
    } else {
      // Fallback for older browsers
      const textArea = document.createElement("textarea");
      textArea.value = content;
      textArea.style.position = "absolute";
      textArea.style.opacity = "0";
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();

      if (document.execCommand("copy")) {
        resolve(content);
      } else {
        reject();
      }

      textArea.remove();
    }
  });

  try {
    await copyPromise;
    toast.success("Copied to clipboard");
  } catch (e) {
    toast.error("Failed to copy to clipboard");
  }
}
