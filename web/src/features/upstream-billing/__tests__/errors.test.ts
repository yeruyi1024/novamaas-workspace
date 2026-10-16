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

import { getServerErrorMessageKey } from '@/lib/server-error-message'

test.each([
  [
    'INVALID_ATTACHMENT',
    'Choose a PDF, XLSX, CSV, PNG or JPEG attachment, no larger than 20 MiB.',
  ],
  ['ATTACHMENT_LIMIT', 'A statement can have at most 20 supplier attachments.'],
  [
    'UPLOAD_PENDING',
    'Attachment uploads are in progress. Wait for them to finish before confirming.',
  ],
  [
    'REVIEW_FORBIDDEN',
    'Financial management permission is required to review statements or upload supplier evidence.',
  ],
  ['FORBIDDEN', 'Access denied'],
  [
    'SESSION_REQUIRED',
    'Sign in with a browser session to confirm this statement.',
  ],
  ['CONFLICT', 'Billing data changed. Refresh and try again.'],
  ['INTEGRITY_FAILED', 'Billing data changed. Refresh and try again.'],
  [
    'INCOMPLETE',
    'Statement preparation failed. Refresh and try again.',
  ],
  [
    'MONTH_OPEN',
    'Select a month that has been closed for at least 24 hours to create an archived draft.',
  ],
  ['NOT_FOUND', 'Content not found.'],
  ['TIMEOUT', 'Please try again later.'],
  ['OPERATION_FAILED', 'Something went wrong!'],
])(
  'upstream %s errors show a translated recovery message without raw server details',
  (code, message) => {
    expect(
      getServerErrorMessageKey({
        response: {
          data: {
            code: `UPSTREAM_BILLING_${code}`,
            message: 'Upstream billing operation could not be completed.',
          },
        },
      })
    ).toBe(message)
  }
)
