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

import {
  assessmentGroup,
  displayMeasured,
  displayThreshold,
  overallLabel,
  VERDICT_LABEL,
  type Assessment,
} from './baselines'
import { PROTOCOL_BASIC_IDS, SHALLOW_BASIC_IDS } from './constants'
import type { CheckResult, CheckStatus } from './types'

export type ReportInput = {
  baseUrl: string
  model: string
  basicChecks: CheckResult[]
  cacheChecks: CheckResult[]
  summaries: { basic: string; cache: string; stress: string }
  stressAssessment: Assessment | null
  cacheAssessment: Assessment | null
  errorMessage: string
  standardLabel: string
  statusLabel: (status: CheckStatus) => string
  t: (key: string, options?: Record<string, string | number>) => string
}

export function stampFileName(): string {
  const now = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}-${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`
}

export function downloadFile(filename: string, content: string, mime: string) {
  const blob = new Blob([content], { type: `${mime};charset=utf-8` })
  const href = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = href
  link.download = filename
  link.click()
  URL.revokeObjectURL(href)
}

export function exportPdfReport(input: ReportInput) {
  const html = buildHtmlReport(input)
  const iframe = document.createElement('iframe')
  iframe.style.position = 'fixed'
  iframe.style.right = '0'
  iframe.style.bottom = '0'
  iframe.style.width = '0'
  iframe.style.height = '0'
  iframe.style.border = '0'
  iframe.title = 'supplier-test-report-print'
  document.body.appendChild(iframe)

  const doc = iframe.contentWindow?.document
  if (!doc) {
    document.body.removeChild(iframe)
    return
  }

  doc.open()
  doc.write(html)
  doc.close()

  iframe.contentWindow?.focus()
  setTimeout(() => {
    iframe.contentWindow?.print()
    setTimeout(() => {
      document.body.removeChild(iframe)
    }, 1000)
  }, 250)
}

function translate(
  t: ReportInput['t'],
  key: string,
  options?: Record<string, string | number>
): string {
  const value = t(key, options)
  return value || key
}

function checkLines(
  checks: CheckResult[],
  input: ReportInput,
  t: ReportInput['t']
): string[] {
  return checks.map((check) => {
    const detail = check.message ? ` — ${check.message}` : ''
    return `- ${t(check.title)}: ${t(input.statusLabel(check.status))}${detail}`
  })
}

function markdownTable(assessment: Assessment, t: ReportInput['t']): string[] {
  if (assessment.rows.length === 0) return []
  const lines = [
    `**${t(overallLabel(assessment.overall))}**`,
    '',
    `| ${t('Metric')} | ${t('Measured')} | ${t('Threshold')} | ${t('Verdict')} |`,
    '|---|---|---|---|',
  ]
  for (const row of assessment.rows) {
    lines.push(
      `| ${t(row.label)} | ${displayMeasured(row, t)} | ${displayThreshold(row, t)} | ${t(VERDICT_LABEL[row.verdict])} |`
    )
  }
  return lines
}

export function buildMarkdownReport(input: ReportInput): string {
  const t: ReportInput['t'] = (key, options) =>
    translate(input.t, key, options)
  const shallowChecks = input.basicChecks.filter((check) =>
    (SHALLOW_BASIC_IDS as readonly string[]).includes(check.id)
  )
  const protocolChecks = input.basicChecks.filter((check) =>
    (PROTOCOL_BASIC_IDS as readonly string[]).includes(check.id)
  )
  const stressShallow = input.stressAssessment
    ? assessmentGroup(input.stressAssessment, 'shallow')
    : null
  const stressPerf = input.stressAssessment
    ? assessmentGroup(input.stressAssessment, 'perf')
    : null

  const lines = [
    `# ${t('Supplier Test Report')}`,
    '',
    `- ${t('Time')}: ${new Date().toLocaleString()}`,
    `- ${t('Base URL')}: ${input.baseUrl || '-'}`,
    `- ${t('Model')}: ${input.model || '-'}`,
    `- ${t('Judgment standard')}: ${input.standardLabel}`,
    '',
    `## ${t('Shallow · connectivity')}`,
    ...checkLines(shallowChecks, input, t),
  ]
  if (input.summaries.basic) {
    lines.push('', input.summaries.basic)
  }
  if (stressShallow && stressShallow.rows.length > 0) {
    lines.push('', ...markdownTable(stressShallow, t))
  }

  lines.push('', `## ${t('Deep · performance')}`)
  lines.push(...checkLines(input.cacheChecks, input, t))
  if (input.cacheAssessment && input.cacheAssessment.rows.length > 0) {
    lines.push('', ...markdownTable(input.cacheAssessment, t))
  }
  if (input.summaries.cache) {
    lines.push('', input.summaries.cache)
  }
  if (stressPerf && stressPerf.rows.length > 0) {
    lines.push('', ...markdownTable(stressPerf, t))
  }
  if (input.summaries.stress) {
    lines.push('', input.summaries.stress)
  }

  lines.push('', `## ${t('Deep · protocol')}`)
  lines.push(...checkLines(protocolChecks, input, t))
  lines.push(
    '',
    t(
      'Protocol checks are skipped when the vendor has no matching API. That is incomplete protocol, not a broken supplier.'
    )
  )

  if (input.errorMessage) {
    lines.push('', `## ${t('Error')}`, input.errorMessage)
  }
  return `${lines.join('\n')}\n`
}

function escapeHtml(value: string): string {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
}

function verdictHtmlColor(verdict: Assessment['overall']): string {
  if (verdict === 'ok') return '#15803d'
  if (verdict === 'slow') return '#a16207'
  if (verdict === 'abnormal') return '#dc2626'
  return '#6b7280'
}

function htmlRows(
  assessment: Assessment,
  t: ReportInput['t']
): string {
  const body = assessment.rows
    .map((row) => {
      const color = verdictHtmlColor(row.verdict)
      return `<tr>
<td>${escapeHtml(t(row.label))}</td>
<td>${escapeHtml(displayMeasured(row, t))}</td>
<td>${escapeHtml(displayThreshold(row, t))}</td>
<td style="color:${color};font-weight:600">${escapeHtml(t(VERDICT_LABEL[row.verdict]))}</td>
</tr>`
    })
    .join('')
  return `<p><strong>${escapeHtml(t(overallLabel(assessment.overall)))}</strong></p>
<table>
<thead><tr><th>${escapeHtml(t('Metric'))}</th><th>${escapeHtml(t('Measured'))}</th><th>${escapeHtml(t('Threshold'))}</th><th>${escapeHtml(t('Verdict'))}</th></tr></thead>
<tbody>${body}</tbody>
</table>`
}

export function buildHtmlReport(input: ReportInput): string {
  const t: ReportInput['t'] = (key, options) =>
    translate(input.t, key, options)
  const checks = (items: CheckResult[]) =>
    items
      .map((check) => {
        const detail = check.message ? ` — ${escapeHtml(check.message)}` : ''
        return `<li>${escapeHtml(t(check.title))}: ${escapeHtml(t(input.statusLabel(check.status)))}${detail}</li>`
      })
      .join('')

  const shallowChecks = input.basicChecks.filter((check) =>
    (SHALLOW_BASIC_IDS as readonly string[]).includes(check.id)
  )
  const protocolChecks = input.basicChecks.filter((check) =>
    (PROTOCOL_BASIC_IDS as readonly string[]).includes(check.id)
  )
  const cacheTable =
    input.cacheAssessment && input.cacheAssessment.rows.length > 0
      ? htmlRows(input.cacheAssessment, t)
      : ''
  const stressShallow =
    input.stressAssessment &&
    assessmentGroup(input.stressAssessment, 'shallow').rows.length > 0
      ? htmlRows(assessmentGroup(input.stressAssessment, 'shallow'), t)
      : ''
  const stressPerf =
    input.stressAssessment &&
    assessmentGroup(input.stressAssessment, 'perf').rows.length > 0
      ? htmlRows(assessmentGroup(input.stressAssessment, 'perf'), t)
      : ''

  return `<!doctype html>
<html lang="zh">
<head>
<meta charset="utf-8"/>
<title>${escapeHtml(t('Supplier Test Report'))}-${stampFileName()}</title>
<style>
@page {
  size: A4;
  margin: 15mm;
}
@media print {
  body {
    -webkit-print-color-adjust: exact;
    print-color-adjust: exact;
  }
}
body{font-family:system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;max-width:960px;margin:24px auto;padding:0 16px;color:#111;line-height:1.6}
h1{font-size:22px;border-bottom:2px solid #2563eb;padding-bottom:8px;margin-bottom:12px;color:#1e293b}
h2{font-size:16px;color:#334155;border-left:4px solid #3b82f6;padding-left:8px;margin-top:24px;margin-bottom:8px}
table{border-collapse:collapse;width:100%;margin:12px 0 20px;font-size:13px}
th,td{border:1px solid #d1d5db;padding:8px 10px;text-align:left;vertical-align:top}
th{background:#f3f4f6;font-weight:600}
.muted{color:#6b7280;font-size:13px}
</style>
</head>
<body>
<h1>${escapeHtml(t('Supplier Test Report'))}</h1>
<p>${escapeHtml(t('Time'))}: ${escapeHtml(new Date().toLocaleString())}<br/>
${escapeHtml(t('Base URL'))}: ${escapeHtml(input.baseUrl || '-')}<br/>
${escapeHtml(t('Model'))}: ${escapeHtml(input.model || '-')}<br/>
${escapeHtml(t('Judgment standard'))}: ${escapeHtml(input.standardLabel)}</p>
<h2>${escapeHtml(t('Shallow · connectivity'))}</h2>
<ul>${checks(shallowChecks)}</ul>
${input.summaries.basic ? `<p class="muted">${escapeHtml(input.summaries.basic)}</p>` : ''}
${stressShallow}
<h2>${escapeHtml(t('Deep · performance'))}</h2>
<ul>${checks(input.cacheChecks)}</ul>
${cacheTable}
${input.summaries.cache ? `<p class="muted">${escapeHtml(input.summaries.cache)}</p>` : ''}
${stressPerf}
${input.summaries.stress ? `<p class="muted">${escapeHtml(input.summaries.stress)}</p>` : ''}
<h2>${escapeHtml(t('Deep · protocol'))}</h2>
<ul>${checks(protocolChecks)}</ul>
<p class="muted">${escapeHtml(t('Protocol checks are skipped when the vendor has no matching API. That is incomplete protocol, not a broken supplier.'))}</p>
${input.errorMessage ? `<h2>${escapeHtml(t('Error'))}</h2><p>${escapeHtml(input.errorMessage)}</p>` : ''}
</body>
</html>
`
}
