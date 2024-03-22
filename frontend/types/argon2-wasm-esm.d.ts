declare module "argon2-wasm-esm" {
  export interface Argon2Options {
    pass: string;
    salt: Uint8Array;
    time: number;
    mem: number;
    hashLen: number;
    parallelism: number;
    type: number;
  }

  export interface Argon2Result {
    hash: Uint8Array;
    hashHex: string;
    encoded: string;
  }

  export function hash(options: Argon2Options): Promise<Argon2Result>;
}
