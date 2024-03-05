import { hash } from "argon2-wasm-esm";

async function deriveKey(
  password: string,
  salt?: Uint8Array
): Promise<[CryptoKey, Uint8Array]> {
  if (!salt) {
    salt = new Uint8Array(16);
    window.crypto.getRandomValues(salt);
  }

  const argon2Options = {
    type: 1, // use Argon2i
    hashLen: 32,
    time: 3,
    mem: 1024 * 32,
    parallelism: 1,
    salt: salt,
    pass: password,
  };

  try {
    const result = await hash(argon2Options);

    const key = await window.crypto.subtle.importKey(
      "raw",
      result.hash,
      { name: "AES-GCM" },
      false,
      ["encrypt", "decrypt"]
    );

    return [key, salt];
  } catch (err) {
    console.error(err);
    throw new Error("Failed to derive key with Argon2");
  }
}

export async function aesGcmEncrypt(
  plaintext: string,
  password: string
): Promise<string> {
  const [secretKey, salt] = await deriveKey(password);
  const plaintextBytes = new TextEncoder().encode(plaintext);

  const iv = new Uint8Array(12);
  window.crypto.getRandomValues(iv);

  const cipherText = await window.crypto.subtle.encrypt(
    {
      name: "AES-GCM",
      iv: iv,
    },
    secretKey,
    plaintextBytes
  );

  return `${arrayBufferToHex(salt)}-${arrayBufferToHex(iv)}-${arrayBufferToHex(
    new Uint8Array(cipherText)
  )}`;
}

export async function aesGcmDecrypt(
  ciphertext: string,
  password: string
): Promise<string> {
  const arr = ciphertext.split("-");
  const salt = hexToArrayBuffer(arr[0]);
  const iv = hexToArrayBuffer(arr[1]);
  const data = hexToArrayBuffer(arr[2]);

  const [secretKey] = await deriveKey(password, salt);

  const plaintext = await window.crypto.subtle.decrypt(
    {
      name: "AES-GCM",
      iv: iv,
    },
    secretKey,
    data
  );

  return new TextDecoder().decode(plaintext);
}

function arrayBufferToHex(buffer: Uint8Array): string {
  return Array.from(buffer)
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");
}

function hexToArrayBuffer(hexString: string): Uint8Array {
  const bytes = new Uint8Array(hexString.length / 2);
  for (let i = 0; i < hexString.length; i += 2) {
    bytes[i / 2] = parseInt(hexString.substr(i, 2), 16);
  }
  return bytes;
}
