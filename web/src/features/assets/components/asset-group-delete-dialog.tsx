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
import { useTranslation } from 'react-i18next'

import { AssetConfirmDialog } from './asset-confirm-dialog'

/**
 * Confirms deleting a group. A group that still holds assets needs an explicit
 * acknowledgement that every asset inside will be deleted with it.
 */
export function AssetGroupDeleteDialog(props: {
  open: boolean
  groupName: string
  /** Undefined when the count is unknown; treated as possibly non-empty. */
  assetCount?: number
  pending: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (withAssets: boolean) => void
}) {
  const { t } = useTranslation()
  const empty = props.assetCount === 0
  let acknowledgeLabel: string | undefined
  if (!empty) {
    acknowledgeLabel =
      props.assetCount === undefined
        ? t('Also delete every asset in this group.')
        : t('Also delete the {{count}} assets in this group.', {
            count: props.assetCount,
          })
  }

  return (
    <AssetConfirmDialog
      open={props.open}
      title={t('Delete group')}
      description={t(
        'Are you sure you want to delete group "{{name}}"? This action cannot be undone.',
        { name: props.groupName }
      )}
      acknowledgeLabel={acknowledgeLabel}
      checkboxId='asset-group-delete-with-assets'
      pending={props.pending}
      onOpenChange={props.onOpenChange}
      onConfirm={() => props.onConfirm(!empty)}
    />
  )
}
