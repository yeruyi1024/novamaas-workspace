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
    return 'Using GLM fields: system-prefix cache and prompt_tokens_details.cached_tokens. GLM-5.3 uses thinking.type plus reasoning_effort.'
  }
  if (selected === 'kimi') {
    return 'Using Kimi fields: prompt_tokens_details.cached_tokens and reasoning_content. K3 uses reasoning_effort; K2.7 sends no thinking control; K2.6 uses thinking.type. The KVV row is a preflight, not certification.'
  }
  if (selected === 'deepseek') {
    return 'Using DeepSeek fields: prompt_cache_hit_tokens and reasoning_content. V4 uses thinking.type plus reasoning_effort; legacy R1 sends no thinking control.'
  }
  return 'Using generic OpenAI-compatible checks. Missing vendor fields are skipped.'
}
