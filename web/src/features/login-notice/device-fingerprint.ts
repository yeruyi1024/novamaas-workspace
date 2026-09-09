/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
function fallbackFingerprint(source: string): string {
  return Array.from({ length: 8 }, (_, index) => {
    let hash = (2166136261 ^ index) >>> 0
    for (let offset = 0; offset < source.length; offset += 1) {
      hash ^= source.charCodeAt(offset)
      hash = Math.imul(hash, 16777619) >>> 0
    }
    return hash.toString(16).padStart(8, '0')
  }).join('')
}

export async function createDeviceFingerprint(): Promise<string> {
  const source = [
    navigator.userAgent,
    navigator.language,
    Intl.DateTimeFormat().resolvedOptions().timeZone,
    `${screen.width}x${screen.height}x${screen.colorDepth}`,
    String(navigator.hardwareConcurrency || 0),
  ].join('|')

  if (!globalThis.crypto?.subtle) {
    return fallbackFingerprint(source)
  }
  const digest = await globalThis.crypto.subtle.digest(
    'SHA-256',
    new TextEncoder().encode(source)
  )
  return Array.from(new Uint8Array(digest), (byte) =>
    byte.toString(16).padStart(2, '0')
  ).join('')
}
