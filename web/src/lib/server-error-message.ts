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
const serverErrorMessageKeys = {
  BILLING_OPERATION_FAILED: 'Something went wrong!',
  BILLING_NOT_FOUND: 'Content not found.',
  BILLING_CONFLICT: 'Billing data changed. Refresh and try again.',
  BILLING_HISTORY_UNRECONCILED:
    'Historical usage exists without formal entries. Review historical accounting before issuing a statement.',
  BILLING_HISTORY_BLOCKED:
    'Historical import is blocked. Refresh the review and resolve all checks before confirming.',
  BILLING_BRANDING_INVALID:
    'Unable to prepare PDF branding. Check the system Logo URL and Footer settings, then try again.',
  BILLING_NOT_CONFIGURED:
    'Ask an administrator to enable formal accounting for this account.',
  BILLING_MONTH_OPEN:
    'Prepare after the month closes plus 24 hours. Review the archived draft before issuing it to the customer.',
  BILLING_PENDING:
    'Unfinished settlements exist. Reconcile them before preparing a statement.',
  BILLING_IDENTITY_REQUIRED: 'Company title and tax ID are required.',
  BILLING_EXISTS: 'An active statement already exists for this month.',
  BILLING_START_LOCKED:
    'Accounting cannot be backdated or changed after posting starts.',
  AUTH_SESSION_LIMIT:
    'Too many active login sessions. On a device where you are already signed in, open Login sessions and use “Sign out other sessions” to revoke them. If you cannot access a signed-in device, reset your password to sign out all sessions.',
  AUTH_SESSION_ISSUANCE_LIMIT:
    'Too many login sessions were created recently. Please wait for the rolling window to pass, then try again.',
  TELEGRAM_BIND_DISABLED: 'Telegram binding is disabled.',
  TELEGRAM_BIND_INVALID_REQUEST:
    'The Telegram authorization request is invalid or expired.',
  TELEGRAM_BIND_FLOW_INVALID:
    'This Telegram binding request has expired or has already been used.',
  TELEGRAM_BIND_SESSION_INVALID:
    'The login session that started this Telegram binding is no longer valid.',
  TELEGRAM_BIND_ALREADY_BOUND: 'This Telegram account is already bound.',
  TELEGRAM_BIND_USER_DELETED: 'This user account no longer exists.',
  TELEGRAM_BIND_USER_DISABLED: 'This user account is disabled.',
  TELEGRAM_BIND_INTERNAL_ERROR: 'Telegram binding failed. Please try again.',
} as const

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object'
}

function serverErrorPayload(value: unknown): Record<string, unknown> | null {
  if (!isRecord(value)) return null

  const response = value.response
  if (isRecord(response) && isRecord(response.data)) {
    return response.data
  }
  return value
}

export function getServerErrorMessageKey(value: unknown): string | null {
  const payload = serverErrorPayload(value)
  if (!payload || typeof payload.code !== 'string') return null

  return (
    serverErrorMessageKeys[
      payload.code as keyof typeof serverErrorMessageKeys
    ] ?? null
  )
}
