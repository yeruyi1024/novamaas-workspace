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
import { CloudUploadIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { AnimatePresence, motion } from 'motion/react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

function carriesFiles(event: DragEvent): boolean {
  return [...(event.dataTransfer?.types ?? [])].includes('Files')
}

/**
 * Full-page drop target: dragging files anywhere over the library shows the
 * overlay, and dropping hands the files to the upload dialog.
 */
export function AssetDropOverlay(props: {
  enabled: boolean
  groupName?: string
  onFiles: (files: File[]) => void
}) {
  const { t } = useTranslation()
  const [active, setActive] = useState(false)
  const depth = useRef(0)
  const onFiles = useRef(props.onFiles)
  useEffect(() => {
    onFiles.current = props.onFiles
  }, [props.onFiles])

  useEffect(() => {
    if (!props.enabled) return
    const enter = (event: DragEvent) => {
      if (!carriesFiles(event)) return
      event.preventDefault()
      depth.current += 1
      setActive(true)
    }
    const over = (event: DragEvent) => {
      if (carriesFiles(event)) event.preventDefault()
    }
    const leave = (event: DragEvent) => {
      if (!carriesFiles(event)) return
      depth.current = Math.max(0, depth.current - 1)
      if (depth.current === 0) setActive(false)
    }
    const drop = (event: DragEvent) => {
      if (!carriesFiles(event)) return
      event.preventDefault()
      depth.current = 0
      setActive(false)
      const files = [...(event.dataTransfer?.files ?? [])]
      if (files.length > 0) onFiles.current(files)
    }
    window.addEventListener('dragenter', enter)
    window.addEventListener('dragover', over)
    window.addEventListener('dragleave', leave)
    window.addEventListener('drop', drop)
    return () => {
      depth.current = 0
      setActive(false)
      window.removeEventListener('dragenter', enter)
      window.removeEventListener('dragover', over)
      window.removeEventListener('dragleave', leave)
      window.removeEventListener('drop', drop)
    }
  }, [props.enabled])

  return (
    <AnimatePresence>
      {props.enabled && active && (
        <motion.div
          aria-hidden='true'
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
          className='bg-background/70 pointer-events-none fixed inset-0 z-50 grid place-items-center p-6 backdrop-blur-md'
        >
          <motion.div
            initial={{ scale: 0.94 }}
            animate={{ scale: 1 }}
            className='border-primary/60 bg-primary/5 flex size-full max-h-[32rem] max-w-3xl flex-col items-center justify-center gap-4 rounded-3xl border-2 border-dashed text-center'
          >
            <span className='bg-primary/10 text-primary grid size-16 place-items-center rounded-2xl'>
              <HugeiconsIcon icon={CloudUploadIcon} className='size-8' />
            </span>
            <div className='space-y-1'>
              <p className='text-lg font-semibold'>
                {props.groupName
                  ? t('Drop to upload to "{{name}}"', { name: props.groupName })
                  : t('Drop to upload')}
              </p>
              <p className='text-muted-foreground text-sm'>
                {t('Images, videos, and audio are supported')}
              </p>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
