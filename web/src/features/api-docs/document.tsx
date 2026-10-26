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
import { createLink, Link, useLocation } from '@tanstack/react-router'
import DOMPurify from 'dompurify'
import { lexer, type MarkedToken } from 'marked'
import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type ComponentPropsWithRef,
} from 'react'
import { useTranslation } from 'react-i18next'
import { code } from 'yace/highlighters/code'

import { CopyButton } from '@/components/copy-button'
import { Markdown } from '@/components/ui/markdown'
import { cn } from '@/lib/utils'

const highlighters: Partial<Record<string, (source: string) => string>> = {
  json: code([
    { type: 'key', pattern: /"(?:\\.|[^"\\])*"(?=\s*:)/ },
    { type: 'num', pattern: /-?\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b/ },
  ]),
  bash: code([
    { type: 'cmt', pattern: /#.*/ },
    { type: 'kw', pattern: /\b(curl|export|echo|tr|base64|source|cat)\b/ },
    { type: 'var', pattern: /--?[a-zA-Z0-9_-]+|\$[A-Z0-9_]+/ },
  ]),
}

const DirectoryLink = createLink(
  (props: ComponentPropsWithRef<'a'> & { current: boolean }) => {
    const { current, ...linkProps } = props
    return <a {...linkProps} aria-current={current ? 'page' : undefined} />
  }
)

export function Document(props: { content: string; section?: 'admin' }) {
  const { t } = useTranslation()
  const hash = useLocation({ select: (location) => location.hash })
  const blocks = useMemo(() => {
    const counts = new Map<string, number>()
    return (lexer(props.content) as MarkedToken[]).map((token, position) => {
      const language = token.type === 'code' ? token.lang?.toLowerCase() : ''
      const highlighter =
        highlighters[language === 'sh' ? 'bash' : (language ?? '')]

      let anchor = ''
      if (token.type === 'heading') {
        const base = token.text
          .toLowerCase()
          .replaceAll(/[^\p{L}\p{N}]+/gu, '-')
          .replaceAll(/^-|-$/g, '')
        const count = counts.get(base) ?? 0
        counts.set(base, count + 1)
        anchor = count > 0 ? `${base}-${count}` : base
      }

      return {
        token,
        position,
        highlighted:
          highlighter && token.type === 'code'
            ? DOMPurify.sanitize(highlighter(token.text), {
                ALLOWED_TAGS: ['span'],
                ALLOWED_ATTR: ['class'],
              })
            : undefined,
        anchor,
      }
    })
  }, [props.content])
  const bodyRef = useRef<HTMLDivElement>(null)
  const directoryRef = useRef<HTMLDivElement>(null)
  const [activeAnchor, setActiveAnchor] = useState('')
  const headings = useMemo(
    () =>
      blocks.filter(
        ({ token }) => token.type === 'heading' && token.depth <= 3
      ),
    [blocks]
  )
  const anchor = blocks.find(
    (block) =>
      block.anchor &&
      (block.anchor === hash || encodeURIComponent(block.anchor) === hash)
  )?.anchor

  useEffect(() => {
    setActiveAnchor(anchor ?? headings[0]?.anchor ?? '')
    if (anchor) document.getElementById(anchor)?.scrollIntoView()

    const elements = headings.flatMap(({ anchor }) => {
      const element = document.getElementById(anchor)
      return element && bodyRef.current?.contains(element) ? [element] : []
    })
    const offset = elements[0]
      ? Number.parseFloat(getComputedStyle(elements[0]).scrollMarginTop) || 112
      : 112
    let frame = 0
    const update = () => {
      frame = 0
      let current: HTMLElement | undefined = elements[0]
      for (const element of elements) {
        const bounds = element.getBoundingClientRect()
        if (!bounds.height) return
        if (bounds.top > offset + 1) break
        current = element
      }
      if (
        window.scrollY > 0 &&
        bodyRef.current &&
        bodyRef.current.getBoundingClientRect().bottom <= window.innerHeight
      ) {
        current = elements.at(-1)
      }
      if (current) setActiveAnchor(current.id)
    }
    const schedule = () => {
      if (!frame) frame = requestAnimationFrame(update)
    }
    window.addEventListener('scroll', schedule, { passive: true })
    window.addEventListener('resize', schedule)
    if (!anchor) schedule()
    return () => {
      window.removeEventListener('scroll', schedule)
      window.removeEventListener('resize', schedule)
      cancelAnimationFrame(frame)
    }
  }, [anchor, headings])

  useEffect(() => {
    const directory = directoryRef.current
    const current = directory?.querySelector('[aria-current="page"]')
    if (!directory || !current) return
    const bounds = directory.getBoundingClientRect()
    const item = current.getBoundingClientRect()
    if (item.top < bounds.top) {
      directory.scrollTop += item.top - bounds.top
    } else if (item.bottom > bounds.bottom) {
      directory.scrollTop += item.bottom - bounds.bottom
    }
  }, [activeAnchor])

  return (
    <div className='grid min-w-0 gap-10 xl:grid-cols-[minmax(0,1fr)_13.5rem]'>
      <div
        ref={bodyRef}
        className='docs-prose flex max-w-[54rem] min-w-0 flex-col gap-6 text-sm leading-7'
      >
        {blocks.map(({ token, anchor, position, highlighted }) => {
          if (token.type === 'space') return null
          if (token.type === 'heading') {
            const isH3 = token.depth === 3
            const Heading = isH3 ? 'h3' : 'h2'
            return (
              <Heading
                id={anchor}
                key={position}
                className={cn(
                  'scroll-mt-28 font-bold tracking-tight text-foreground',
                  isH3
                    ? 'mt-6 text-base font-semibold'
                    : 'mt-10 pb-2.5 text-2xl border-b border-border/50 first:mt-0'
                )}
              >
                <Link
                  to='/docs'
                  search={(previous) => previous}
                  hash={anchor}
                  onClick={() => setActiveAnchor(anchor)}
                  className='group hover:text-primary focus-visible:outline-ring inline-flex w-fit items-center gap-2 rounded-sm outline-offset-4 transition-colors'
                >
                  <span>{token.text}</span>
                  <span
                    aria-hidden
                    className='text-primary/70 font-normal opacity-0 transition-opacity select-none group-hover:opacity-100 group-focus-visible:opacity-100'
                  >
                    #
                  </span>
                </Link>
              </Heading>
            )
          }
          if (token.type === 'code') {
            return (
              <div
                key={position}
                className='docs-code border-border/70 my-3 min-w-0 overflow-hidden rounded-xl border'
              >
                <div className='docs-code-header flex items-center justify-between gap-3 border-b px-4 py-2.5'>
                  <div className='flex items-center gap-2.5'>
                    <div className='flex items-center gap-1.5' aria-hidden>
                      <span className='inline-block size-2.5 rounded-full bg-rose-500/70' />
                      <span className='inline-block size-2.5 rounded-full bg-amber-500/70' />
                      <span className='inline-block size-2.5 rounded-full bg-emerald-500/70' />
                    </div>
                    <span className='text-muted-foreground pl-1 font-mono text-[11px] font-semibold tracking-wider uppercase'>
                      {token.lang && /^(bash|sh)$/i.test(token.lang)
                        ? 'CURL'
                        : token.lang?.toUpperCase() || 'TEXT'}
                    </span>
                  </div>
                  <CopyButton value={token.text} size='sm'>
                    {t('Copy')}
                  </CopyButton>
                </div>
                <pre
                  className='overflow-x-auto p-4 font-mono text-[13px] leading-6 outline-offset-[-3px]'
                  tabIndex={0}
                  aria-label={t('Code example')}
                >
                  {highlighted ? (
                    <code dangerouslySetInnerHTML={{ __html: highlighted }} />
                  ) : (
                    <code>{token.text}</code>
                  )}
                </pre>
              </div>
            )
          }
          if (token.type === 'table') {
            return (
              <div
                key={position}
                role='region'
                aria-label={
                  props.section === 'admin'
                    ? t('Administrator manual')
                    : t('API documentation')
                }
                tabIndex={0}
                className='docs-table border-border/70 my-3 min-w-0 overflow-x-auto rounded-xl border outline-offset-4'
              >
                <Markdown>
                  {token.raw.replaceAll(
                    /\| *(POST|GET|DELETE|PUT|PATCH) *\|/g,
                    (_, method: string) =>
                      `| <span class="docs-method-badge docs-method-${method.toLowerCase()}">${method}</span> |`
                  )}
                </Markdown>
              </div>
            )
          }
          return (
            <Markdown
              key={position}
              className='[&_blockquote]:bg-primary/[0.04] [&_blockquote]:border-l-primary/70 [&_blockquote]:rounded-r-lg'
            >
              {token.raw}
            </Markdown>
          )
        })}
      </div>
      <nav aria-label={t('On this page')} className='hidden xl:block'>
        <div
          ref={directoryRef}
          className='border-border/60 sticky top-28 flex max-h-[calc(100svh-9rem)] flex-col gap-2.5 overflow-y-auto border-l pl-4'
        >
          <p className='text-muted-foreground mb-1 text-[11px] font-semibold tracking-wider uppercase'>
            {t('On this page')}
          </p>
          {headings.map(({ token, anchor }) => (
            <DirectoryLink
              key={anchor}
              to='/docs'
              search={(previous) => previous}
              hash={anchor}
              activeOptions={{ includeHash: true }}
              activeProps={{}}
              current={activeAnchor === anchor}
              onClick={() => setActiveAnchor(anchor)}
              className={cn(
                'block text-xs leading-5 transition-all',
                activeAnchor === anchor
                  ? 'text-primary font-medium border-l-2 -ml-[17px] pl-3 border-primary'
                  : 'text-muted-foreground hover:text-foreground',
                token.type === 'heading' && token.depth === 3 && 'pl-3'
              )}
            >
              {token.type === 'heading' && token.text}
            </DirectoryLink>
          ))}
        </div>
      </nav>
    </div>
  )
}
