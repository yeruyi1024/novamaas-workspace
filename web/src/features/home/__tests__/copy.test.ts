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
import { expect, test } from 'vitest'

import en from '@/i18n/locales/en.json'
import fr from '@/i18n/locales/fr.json'
import ja from '@/i18n/locales/ja.json'
import ru from '@/i18n/locales/ru.json'
import vi from '@/i18n/locales/vi.json'
import zhTW from '@/i18n/locales/zh-TW.json'
import zh from '@/i18n/locales/zh.json'

const homepageTitleKeys = [
  'Access the right supply through one endpoint',
  'Build a distribution business with control',
  'Build the operating system for AI supply',
  'One platform, multiple business models',
  'One platform, three layers of value',
  'one programmable market',
  'Operate the economics, not just the API',
  'Turn fragmented inventory into reachable demand',
] as const

const localeTranslations = [
  ['en', en.translation],
  ['zh', zh.translation],
  ['zh-TW', zhTW.translation],
  ['fr', fr.translation],
  ['ja', ja.translation],
  ['ru', ru.translation],
  ['vi', vi.translation],
] as const

test.each(localeTranslations)(
  '%s homepage titles do not end with sentence punctuation',
  (_locale, translation) => {
    for (const key of homepageTitleKeys) {
      expect(translation[key]).not.toMatch(/[.!?。！？]$/u)
    }
  }
)
