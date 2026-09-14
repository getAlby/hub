import { localStorageKeys } from "src/constants";

/**
 * Alby Hub used to keep the session token in localStorage and send it in the
 * Authorization header. Sessions are now an HttpOnly cookie, so clear out any
 * token left behind by an older version - it stays valid until it expires.
 */
export function deleteLegacyAuthToken() {
  localStorage.removeItem(localStorageKeys.authToken);
}
