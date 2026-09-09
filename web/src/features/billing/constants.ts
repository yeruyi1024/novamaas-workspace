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
export const statementStatusKeys = {
  preparing: 'Preparing',
  draft: 'Draft',
  issued: 'Awaiting confirmation',
  disputed: 'Disputed',
  confirmed: 'Confirmed',
  void: 'Voided',
  failed: 'Failed',
} as const
export const statementActionKeys: Record<string, string> = {
  prepare: 'Create archived draft',
  issue: 'Issue to customer',
  confirm: 'Confirm statement',
  dispute: 'Raise a dispute',
  reply: 'Reply',
  void: 'Void statement',
  retry: 'Retry',
}
