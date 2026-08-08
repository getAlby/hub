#!/usr/bin/env python3
"""
Extract Greenlight device credentials into PEM files the Alby hub can use.

Greenlight stores device credentials as a single protobuf blob (gl-client's
credentials::Device serialized via prost):

    field 1: version   (uint32, varint)
    field 2: cert      (bytes, PEM certificate)
    field 3: key       (bytes, PEM private key)
    field 4: ca        (bytes, PEM CA certificate)
    field 5: rune      (string)

Default location: ~/.local/share/greenlight/creds  (gl-client DataDir)

Usage:
    python3 extract_creds.py <creds-blob> <output-dir>

Writes ca.pem, client.pem, client-key.pem, rune into <output-dir> and prints
the GREENLIGHT_* environment variables to configure the hub backend.
"""
import base64
import os
import re
import subprocess
import sys


def read_varint(data: bytes, pos: int) -> tuple[int, int]:
    result = 0
    shift = 0
    while True:
        if pos >= len(data):
            raise ValueError("truncated varint")
        b = data[pos]
        pos += 1
        result |= (b & 0x7F) << shift
        if not b & 0x80:
            return result, pos
        shift += 7


def parse_creds(data: bytes) -> dict:
    """Decode the Device protobuf. Returns {1: version, 2: cert, 3: key, 4: ca, 5: rune}."""
    fields = {}
    pos = 0
    while pos < len(data):
        key, pos = read_varint(data, pos)
        field_num = key >> 3
        wire_type = key & 0x07
        if wire_type == 0:
            value, pos = read_varint(data, pos)
            fields[field_num] = value
        elif wire_type == 2:
            length, pos = read_varint(data, pos)
            if pos + length > len(data):
                raise ValueError("truncated length-delimited field")
            fields[field_num] = data[pos : pos + length]
            pos += length
        elif wire_type == 1:
            pos += 8
        elif wire_type == 5:
            pos += 4
        else:
            raise ValueError(f"unsupported wire type {wire_type} at offset {pos}")
    return fields


_CHARSET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"


def _bech32_polymod(values: list[int]) -> int:
    gen = [0x3B6A57B2, 0x26508E6D, 0x1EA119FA, 0x3D4233DD, 0x2A1462B3]
    chk = 1
    for v in values:
        b = chk >> 25
        chk = (chk & 0x1FFFFFF) << 5 ^ v
        for i in range(5):
            chk ^= gen[i] if (b >> i) & 1 else 0
    return chk


def _hrp_expand(hrp: str) -> list[int]:
    return [ord(x) >> 5 for x in hrp] + [0] + [ord(x) & 31 for x in hrp]


def _convertbits(data: bytes, frombits: int, tobits: int, pad: bool = True) -> list[int]:
    acc = 0
    bits = 0
    ret = []
    maxv = (1 << tobits) - 1
    for value in data:
        acc = (acc << frombits) | value
        bits += frombits
        while bits >= tobits:
            bits -= tobits
            ret.append((acc >> bits) & maxv)
    if pad and bits:
        ret.append((acc << (tobits - bits)) & maxv)
    return ret


def bech32m_encode(data: bytes, hrp: str = "gl") -> str:
    """bech32m-encode bytes with the given human-readable prefix."""
    data5 = _convertbits(data, 8, 5)
    values = _hrp_expand(hrp) + data5
    polymod = _bech32_polymod(values + [0, 0, 0, 0, 0, 0]) ^ 0x2BC830A3
    checksum = [(polymod >> 5 * (5 - i)) & 31 for i in range(6)]
    return hrp + "1" + "".join(_CHARSET[d] for d in data5 + checksum)


def node_domain(client_cert_path: str) -> str:
    """Derive the stable node-domain URL for a Greenlight node.

    The device cert subject CN is /users/<hex compressed pubkey>/<slot> — or
    /users/<hex compressed pubkey> (no slot). The node domain is a reverse
    proxy whose SNI label is the bech32m encoding of the 33-byte compact
    pubkey: gl1<bech32m>.nodes.gl.blckstrm.com (verified against a live
    production scheduler 2026-08-02).
    """
    out = subprocess.run(
        ["openssl", "x509", "-in", client_cert_path, "-noout", "-subject"],
        capture_output=True,
        text=True,
        check=True,
    )
    subject = out.stdout.strip()
    m = re.search(r"CN\s*=\s*([^,\n]+)", subject)
    if not m:
        raise ValueError(f"no CN in subject: {subject}")
    parts = [p for p in m.group(1).strip().split("/") if p]
    pubkey_hex = next(p for p in parts if len(p) == 66 and all(c in "0123456789abcdefABCDEF" for c in p))
    pubkey = bytes.fromhex(pubkey_hex)
    return f"{bech32m_encode(pubkey)}.nodes.gl.blckstrm.com"


def main() -> int:
    if len(sys.argv) != 3:
        print(__doc__)
        return 2

    creds_path = sys.argv[1]
    out_dir = sys.argv[2]

    if not os.path.isfile(creds_path):
        print(f"error: creds file not found: {creds_path}", file=sys.stderr)
        return 1

    with open(creds_path, "rb") as f:
        data = f.read()

    fields = parse_creds(data)

    cert = fields.get(2)
    key = fields.get(3)
    ca = fields.get(4)
    rune = fields.get(5)

    if not cert or not key or not ca:
        print(
            "error: creds blob missing required fields (cert=%s key=%s ca=%s)"
            % (bool(cert), bool(key), bool(ca)),
            file=sys.stderr,
        )
        return 1

    os.makedirs(out_dir, exist_ok=True)
    os.chmod(out_dir, 0o700)

    files = {
        "client.pem": cert,
        "client-key.pem": key,
        "ca.pem": ca,
        "rune": rune if rune else b"",
    }
    for name, content in files.items():
        with open(os.path.join(out_dir, name), "wb") as f:
            f.write(content)
        os.chmod(os.path.join(out_dir, name), 0o600)

    print(f"wrote credentials to {out_dir}")
    print()

    try:
        domain = node_domain(os.path.join(out_dir, "client.pem"))
        print(f"node domain: {domain}")
        print()
        print("configure the hub backend:")
        print(f"  GREENLIGHT_CREDS_PATH={out_dir}")
        print(f"  GREENLIGHT_NODE_URI={domain}:443")
        print("  LNBackendType=GREENLIGHT")
        print()
        print("(GREENLIGHT_SERVER_NAME is optional; defaults to the node domain host)")
    except Exception as e:  # noqa: BLE001
        print(f"warning: could not derive node domain from client cert: {e}", file=sys.stderr)
        print("set GREENLIGHT_NODE_URI manually to gl1<node_id>.gl.blckstrm.com:443")

    return 0


if __name__ == "__main__":
    sys.exit(main())
