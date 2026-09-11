import assert from "node:assert/strict";
import { test } from "node:test";

import { withAlbyAuthHeader } from "../../../src/lib/authHeaders.ts";

test("preserves Authorization from a Headers instance", () => {
  const headers = new Headers({
    Authorization: "Basic dXNlcjpwYXNzd29yZA==",
  });

  const result = withAlbyAuthHeader(headers, "hub-token");

  assert.equal(result.get("Authorization"), "Basic dXNlcjpwYXNzd29yZA==");
  assert.equal(result.get("X-Alby-Auth"), "Bearer hub-token");
});
