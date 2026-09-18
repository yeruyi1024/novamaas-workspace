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

export type VendorId = 'generic' | 'glm' | 'kimi' | 'deepseek'

export const VENDOR_OPTIONS: Array<{ id: VendorId; labelKey: string }> = [
  { id: 'generic', labelKey: 'Generic OpenAI-compatible' },
  { id: 'glm', labelKey: 'GLM' },
  { id: 'kimi', labelKey: 'Kimi' },
  { id: 'deepseek', labelKey: 'DeepSeek' },
]

export function vendorHintKey(selected: VendorId): string {
  if (selected === 'glm') {
    return 'Using GLM fields: system-prefix cache, reasoning_content, cached_tokens.'
  }
  if (selected === 'kimi') {
    return 'Using Kimi fields: cached_tokens, reasoning_content. kimi-k3 does not send thinking.'
  }
  if (selected === 'deepseek') {
    return 'Using DeepSeek fields: prompt_cache_hit_tokens, reasoning_content. R1 does not send thinking.'
  }
  return 'Using generic OpenAI-compatible checks. Missing vendor fields are skipped.'
}
