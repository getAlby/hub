import { localStorageKeys } from "src/constants";

export function getAuthToken() {
  return localStorage.getItem(localStorageKeys.authToken);
}

export function saveAuthToken(token: string) {
  localStorage.setItem(localStorageKeys.authToken, token);
}
