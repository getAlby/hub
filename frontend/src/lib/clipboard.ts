export function copyToClipboard(content: string) {
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(content);
  } else {
    // Fallback for older browsers
    const textArea = document.createElement("textarea");
    textArea.value = content;
    textArea.style.position = "absolute";
    textArea.style.opacity = "0";
    document.body.appendChild(textArea);
    textArea.select();
    new Promise((res, rej) => {
      document.execCommand("copy") ? res(content) : rej();
      textArea.remove();
    });
  }
}
