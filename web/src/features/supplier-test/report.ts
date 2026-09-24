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
  displayMeasured,
  displayThreshold,
  formatPercent,
  isInformationalRow,
  overallLabel,
  VERDICT_LABEL,
  type Assessment,
} from './baselines'
import { PROTOCOL_BASIC_IDS, SHALLOW_BASIC_IDS } from './constants'
import type {
  CacheMetrics,
  CheckResult,
  CheckStatus,
  StressMetrics,
} from './types'

export type StressConfig = {
  concurrency: number
  rounds: number
  stream: boolean
  breakCache: boolean
  maxTokens?: number
  corpus?: string
}

export type ReportInput = {
  baseUrl: string
  model: string
  vendor?: string
  basicChecks: CheckResult[]
  cacheChecks: CheckResult[]
  summaries: { basic: string; cache: string; stress: string }
  stressAssessment: Assessment | null
  cacheAssessment: Assessment | null
  errorMessage: string
  standardLabel: string
  statusLabel: (status: CheckStatus) => string
  stressConfig?: StressConfig
  stressMetrics?: StressMetrics | null
  cacheMetrics?: CacheMetrics | null
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

  printReportFrame(iframe, doc)
}

function printReportFrame(iframe: HTMLIFrameElement, doc: Document) {
  const print = () => {
    iframe.contentWindow?.focus()
    iframe.contentWindow?.print()
    setTimeout(() => {
      if (iframe.isConnected) document.body.removeChild(iframe)
    }, 1000)
  }
  const fontsReady = doc.fonts?.ready
  if (fontsReady) {
    void fontsReady.then(
      () => setTimeout(print, 100),
      () => setTimeout(print, 250)
    )
  } else {
    setTimeout(print, 250)
  }
}

function translate(
  t: ReportInput['t'],
  key: string,
  options?: Record<string, string | number>
): string {
  const value = t(key, options)
  return value || key
}

export function formatTokenCompact(tokens: number | undefined): string {
  if (tokens === undefined || !Number.isFinite(tokens) || tokens <= 0) {
    return '0'
  }
  if (tokens >= 1_000_000) return `${(tokens / 1_000_000).toFixed(1)}M`
  if (tokens >= 1_000) return `${(tokens / 1_000).toFixed(1)}k`
  return String(tokens)
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
  const benchmarkRows = assessment.rows.filter(
    (row) => !isInformationalRow(row)
  )
  if (benchmarkRows.length === 0) return []
  const lines = [
    `| ${t('Metric')} | ${t('Measured')} | ${t('Threshold')} | ${t('Verdict')} |`,
    '|---|---|---|---|',
  ]
  for (const row of benchmarkRows) {
    lines.push(
      `| ${t(row.label)} | ${displayMeasured(row, t)} | ${displayThreshold(row, t)} | ${t(VERDICT_LABEL[row.verdict])} |`
    )
  }
  return lines
}

export function buildMarkdownReport(input: ReportInput): string {
  const t: ReportInput['t'] = (key, options) => translate(input.t, key, options)
  const shallowChecks = input.basicChecks.filter((check) =>
    (SHALLOW_BASIC_IDS as readonly string[]).includes(check.id)
  )
  const protocolChecks = input.basicChecks.filter((check) => {
    if (check.id === 'kimi_kvv' && input.vendor && input.vendor !== 'kimi') {
      return false
    }
    return (PROTOCOL_BASIC_IDS as readonly string[]).includes(check.id)
  })
  const lines = [
    `# ${t('Supplier Test Report')}`,
    '',
    `- ${t('Time')}: ${new Date().toLocaleString()}`,
    `- ${t('Base URL')}: ${input.baseUrl || '-'}`,
    `- ${t('Model')}: ${input.model || '-'}`,
    `- ${t('Judgment standard')}: ${input.standardLabel}`,
    '',
    `## 1. ${t('Connectivity and protocol')}`,
    `### 1.1 ${t('Basic connectivity checks')}`,
    ...checkLines(shallowChecks, input, t),
    '',
    `### 1.2 ${t('Advanced protocol capabilities')}`,
    ...checkLines(protocolChecks, input, t),
  ]
  if (input.summaries.basic) {
    lines.push('', input.summaries.basic)
  }

  // Section 2: Stress test
  if (input.stressAssessment || input.stressMetrics || input.summaries.stress) {
    lines.push('', `## 2. ${t('Concurrency and stress test')}`)
    if (input.stressAssessment) {
      lines.push(
        `**${t('Stress test assessment')}：${t(overallLabel(input.stressAssessment.overall))}**`,
        ''
      )
    }
    if (input.stressConfig) {
      const cfg = input.stressConfig
      const modeText = cfg.stream ? t('Stream') : t('Non-stream')
      const cacheText = cfg.breakCache ? t('Break cache') : t('Allow cache')
      const totalReq = cfg.concurrency * cfg.rounds
      lines.push(
        `- ${t('Concurrency spec')}: ${cfg.concurrency} × ${cfg.rounds} (${totalReq})`,
        `- ${t('Transmission mode')}: ${modeText}`,
        `- ${t('Cache strategy')}: ${cacheText}`,
        `- ${t('Max tokens cap')}: max_tokens = ${cfg.maxTokens ?? 1024}`,
        ''
      )
    }
    if (input.stressMetrics) {
      const sm = input.stressMetrics
      const completeUsage =
        sm.usage_n === undefined || sm.usage_n === sm.succeeded
      lines.push(
        `- ${t('Success / failed')}: ${sm.succeeded} / ${sm.failed}`,
        `- ${t('Requests completed / not run')}: ${sm.attempted ?? sm.total} / ${sm.not_run ?? 0}`,
        `- ${t('Batch wall time')}: ${(sm.elapsed_ms / 1000).toFixed(2)} s`,
        `- ${t('Batch output throughput')}: ${completeUsage ? `${sm.tokens_per_sec.toFixed(1)} tok/s` : '—'}`,
        `- ${t('Per-request output rate')}: ${completeUsage && sm.request_tokens_per_sec ? `${sm.request_tokens_per_sec.toFixed(1)} tok/s` : '—'}`,
        `- ${t('Request duration avg / P50 / P90')}: ${sm.request_avg_ms ? `${(sm.request_avg_ms / 1000).toFixed(2)} / ${((sm.request_p50_ms ?? 0) / 1000).toFixed(2)} / ${((sm.request_p90_ms ?? 0) / 1000).toFixed(2)} s` : '—'}`,
        `- ${t('Tokens (prompt / completion)')}: ${formatTokenCompact(sm.prompt_tokens)} / ${formatTokenCompact(sm.completion_tokens)}`,
        `- ${t('Estimated capacity')}: ${Math.round(sm.rpm)} RPM / ${completeUsage ? formatTokenCompact(sm.tpm) : '—'} TPM`,
        ''
      )
      const missingUsage = Math.max(
        0,
        sm.succeeded - (sm.usage_n ?? sm.succeeded)
      )
      const notRun =
        sm.not_run ?? Math.max(0, sm.total - (sm.attempted ?? sm.total))
      const missingTTFT = input.stressConfig?.stream
        ? Math.max(0, sm.succeeded - sm.ttft_n)
        : 0
      const missingTPOT = input.stressConfig?.stream
        ? Math.max(0, sm.ttft_n - sm.tpot_n)
        : 0
      const slowRows =
        input.stressAssessment?.rows.filter(
          (row) => row.verdict === 'slow' || row.verdict === 'abnormal'
        ) ?? []
      lines.push(`### ${t('Problems found')}`)
      if (
        sm.failed === 0 &&
        notRun === 0 &&
        missingUsage === 0 &&
        missingTTFT === 0 &&
        missingTPOT === 0 &&
        slowRows.length === 0
      ) {
        lines.push(
          t('No request errors or threshold violations were found in this run.')
        )
      }
      for (const issue of sm.issues ?? []) {
        const status = issue.status_code || t('No HTTP response')
        lines.push(
          `- ${issue.count} × ${status}: ${issue.message.replaceAll('|', '\\|')} (W${issue.worker}/R${issue.round}, ${(issue.elapsed_ms / 1000).toFixed(2)} s)`
        )
      }
      if (sm.other_issue_count) {
        lines.push(
          `- ${t('{{count}} additional failures are included in the failed total but omitted from the detailed list.', { count: sm.other_issue_count })}`
        )
      }
      if (sm.failed > 0 && !sm.issues?.length && !sm.other_issue_count) {
        lines.push(`- ${t('Failure details are unavailable for this run.')}`)
      }
      if (missingUsage > 0) {
        lines.push(
          `- ${t('{{count}} successful requests had no complete token usage; token rates and TPM are unavailable.', { count: missingUsage })}`
        )
      }
      if (notRun > 0) {
        lines.push(
          `- ${t('{{count}} planned requests were not started because the run stopped or was cancelled.', { count: notRun })}`
        )
      }
      if (missingTTFT > 0) {
        lines.push(
          `- ${t('{{count}} successful stream requests had no observable first output; TTFT is unavailable for them.', { count: missingTTFT })}`
        )
      }
      if (missingTPOT > 0) {
        lines.push(
          `- ${t('{{count}} streamed responses lacked enough timed output and token usage to calculate TPOT.', { count: missingTPOT })}`
        )
      }
      for (const row of slowRows) {
        lines.push(
          `- ${t(row.label)}: ${displayMeasured(row, t)} · ${displayThreshold(row, t)}`
        )
      }
      lines.push('')
    }
    if (input.stressAssessment) {
      lines.push(...markdownTable(input.stressAssessment, t))
    }
    if (input.summaries.stress) {
      lines.push('', input.summaries.stress)
    }
  }

  // Section 3: Cache test
  if (input.cacheAssessment || input.cacheMetrics || input.summaries.cache) {
    lines.push('', `## 3. ${t('Prompt cache test')}`)
    if (input.cacheAssessment) {
      lines.push(
        `**${t('Prompt cache assessment')}：${t(overallLabel(input.cacheAssessment.overall))}**`,
        ''
      )
    }
    if (input.cacheChecks.length > 0) {
      lines.push(
        `### 3.1 ${t('Cache probe details')}`,
        ...checkLines(input.cacheChecks, input, t),
        ''
      )
    }
    if (input.cacheMetrics) {
      const cm = input.cacheMetrics
      const modeLabel =
        cm.mode === 'cumulative' ? t('Cumulative chat') : t('Static prefix')
      const freqText =
        cm.hit_count !== undefined
          ? `${cm.hit_count}/${cm.rounds} (${formatPercent(cm.rounds > 0 ? (cm.hit_count ?? 0) / cm.rounds : 0)})`
          : '-'
      lines.push(
        `- ${t('Cache mode')}: ${modeLabel}`,
        `- ${t('Probe rounds')}: ${cm.rounds}`,
        `- ${t('Hit frequency')}: ${freqText}`,
        `- ${t('Hit depth')}: ${cm.avg_depth_rate !== undefined ? formatPercent(cm.avg_depth_rate) : '-'}`,
        `- ${t('TTL wait (s)')}: ${cm.wait_seconds}s`,
        ''
      )
    }
    if (input.cacheAssessment) {
      lines.push(...markdownTable(input.cacheAssessment, t))
    }
    if (input.summaries.cache) {
      lines.push('', input.summaries.cache)
    }
  }

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

function verdictHtmlClass(verdict: Assessment['overall']): string {
  if (verdict === 'ok') return 'verdict-ok'
  if (verdict === 'slow') return 'verdict-slow'
  if (verdict === 'abnormal') return 'verdict-abnormal'
  return 'verdict-na'
}

function renderCheckListHtml(
  checks: CheckResult[],
  input: ReportInput,
  t: ReportInput['t']
): string {
  if (checks.length === 0) return ''
  const items = checks
    .map((check) => {
      let statusCls = 'check-status-fail'
      if (check.status === 'pass') {
        statusCls = 'check-status-pass'
      } else if (check.status === 'skip') {
        statusCls = 'check-status-skip'
      }
      const detail = check.message
        ? ` <span class="check-detail">— ${escapeHtml(check.message)}</span>`
        : ''
      return `<li class="check-list-item">
<span class="check-bullet">•</span>
<span class="check-name">${escapeHtml(t(check.title))}：</span>
<span class="${statusCls}">${escapeHtml(t(input.statusLabel(check.status)))}</span>${detail}
</li>`
    })
    .join('')
  return `<ul class="check-list">${items}</ul>`
}

function htmlBenchmarkTable(
  assessment: Assessment,
  t: ReportInput['t']
): string {
  const benchmarkRows = assessment.rows.filter(
    (row) => !isInformationalRow(row)
  )
  if (benchmarkRows.length === 0) return ''
  const body = benchmarkRows
    .map((row) => {
      const color = verdictHtmlColor(row.verdict)
      return `<tr>
<td>${escapeHtml(t(row.label))}</td>
<td class="cell-measured">${escapeHtml(displayMeasured(row, t))}</td>
<td class="cell-threshold">${escapeHtml(displayThreshold(row, t))}</td>
<td style="color:${color};font-weight:600">${escapeHtml(t(VERDICT_LABEL[row.verdict]))}</td>
</tr>`
    })
    .join('')
  return `<table class="benchmark-table">
<thead><tr><th style="width:25%">${escapeHtml(t('Metric'))}</th><th style="width:20%">${escapeHtml(t('Measured'))}</th><th style="width:40%">${escapeHtml(t('Threshold'))}</th><th style="width:15%">${escapeHtml(t('Verdict'))}</th></tr></thead>
<tbody>${body}</tbody>
</table>`
}

export function buildHtmlReport(input: ReportInput): string {
  const t: ReportInput['t'] = (key, options) => translate(input.t, key, options)

  const shallowChecks = input.basicChecks.filter((check) =>
    (SHALLOW_BASIC_IDS as readonly string[]).includes(check.id)
  )
  const protocolChecks = input.basicChecks.filter((check) => {
    if (check.id === 'kimi_kvv' && input.vendor && input.vendor !== 'kimi') {
      return false
    }
    return (PROTOCOL_BASIC_IDS as readonly string[]).includes(check.id)
  })
  const basicResults = [...shallowChecks, ...protocolChecks]
  let basicBadge = 'All checks passed'
  let basicBadgeClass = 'verdict-ok'
  if (basicResults.some((check) => check.status === 'fail')) {
    basicBadge = 'Checks have failures'
    basicBadgeClass = 'verdict-abnormal'
  } else if (basicResults.some((check) => check.status === 'skip')) {
    basicBadge = 'Checks complete with skips'
    basicBadgeClass = 'verdict-na'
  }

  // Section 2: Stress test HTML
  let stressSection = ''
  if (input.stressAssessment || input.stressMetrics || input.summaries.stress) {
    const overall = input.stressAssessment?.overall ?? 'ok'
    const overallCls = verdictHtmlClass(overall)
    const badgeText = `${t('Stress test assessment')}：${t(overallLabel(overall))}`

    let configBar = ''
    if (input.stressConfig) {
      const cfg = input.stressConfig
      const modeText = cfg.stream ? t('Stream') : t('Non-stream')
      const cacheText = cfg.breakCache ? t('Break cache') : t('Allow cache')
      const total = cfg.concurrency * cfg.rounds
      configBar = `<div class="config-bar">
<div class="config-item">${escapeHtml(t('Concurrency spec'))}：<span>${cfg.concurrency} × ${cfg.rounds} (${total})</span></div>
<div class="config-item">${escapeHtml(t('Transmission mode'))}：<span>${escapeHtml(modeText)}</span></div>
<div class="config-item">${escapeHtml(t('Cache strategy'))}：<span>${escapeHtml(cacheText)}</span></div>
<div class="config-item">${escapeHtml(t('Max tokens cap'))}：<span>max_tokens = ${cfg.maxTokens ?? 1024}</span></div>
</div>`
    }

    let statsGrid = ''
    let problemsHtml = ''
    if (input.stressMetrics) {
      const sm = input.stressMetrics
      const durationSec = (sm.elapsed_ms / 1000).toFixed(2)
      const completeUsage =
        sm.usage_n === undefined || sm.usage_n === sm.succeeded
      statsGrid = `<div class="stats-grid">
<div class="stat-card">
  <div class="stat-val">${sm.succeeded} / ${sm.failed}</div>
  <div class="stat-lbl">${escapeHtml(t('Success / failed'))}</div>
</div>
<div class="stat-card">
  <div class="stat-val">${sm.attempted ?? sm.total} / ${sm.not_run ?? 0}</div>
  <div class="stat-lbl">${escapeHtml(t('Requests completed / not run'))}</div>
</div>
<div class="stat-card">
  <div class="stat-val">${durationSec} s</div>
  <div class="stat-lbl">${escapeHtml(t('Batch wall time'))}</div>
</div>
<div class="stat-card">
  <div class="stat-val">${completeUsage ? `${sm.tokens_per_sec.toFixed(1)} tok/s` : '—'}</div>
  <div class="stat-lbl">${escapeHtml(t('Batch output throughput'))}</div>
</div>
<div class="stat-card">
  <div class="stat-val">${sm.request_avg_ms ? `${(sm.request_avg_ms / 1000).toFixed(2)} / ${((sm.request_p50_ms ?? 0) / 1000).toFixed(2)} / ${((sm.request_p90_ms ?? 0) / 1000).toFixed(2)} s` : '—'}</div>
  <div class="stat-lbl">${escapeHtml(t('Request duration avg / P50 / P90'))}</div>
</div>
<div class="stat-card">
  <div class="stat-val">${completeUsage && sm.request_tokens_per_sec ? `${sm.request_tokens_per_sec.toFixed(1)} tok/s` : '—'}</div>
  <div class="stat-lbl">${escapeHtml(t('Per-request output rate'))}</div>
</div>
<div class="stat-card">
  <div class="stat-val">${sm.ttft_n > 0 ? `${(sm.ttft_avg_ms / 1000).toFixed(2)} / ${(sm.ttft_p50_ms / 1000).toFixed(2)} / ${(sm.ttft_p90_ms / 1000).toFixed(2)} s` : '—'}</div>
  <div class="stat-lbl">${escapeHtml(t('TTFT avg / P50 / P90'))}</div>
</div>
<div class="stat-card">
  <div class="stat-val">${sm.tpot_n > 0 ? `${sm.tpot_avg_ms.toFixed(1)} / ${sm.tpot_p50_ms.toFixed(1)} / ${sm.tpot_p90_ms.toFixed(1)} ms` : '—'}</div>
  <div class="stat-lbl">${escapeHtml(t('TPOT avg / P50 / P90'))}</div>
</div>
<div class="stat-card">
  <div class="stat-val">${formatTokenCompact(sm.prompt_tokens)} / ${formatTokenCompact(sm.completion_tokens)}</div>
  <div class="stat-lbl">${escapeHtml(t('Tokens (prompt / completion)'))}</div>
</div>
<div class="stat-card">
  <div class="stat-val">${Math.round(sm.rpm)} / ${completeUsage ? formatTokenCompact(sm.tpm) : '—'}</div>
  <div class="stat-lbl">${escapeHtml(t('Estimated capacity (RPM/TPM)'))}</div>
</div>
</div>`
      const missingUsage = Math.max(
        0,
        sm.succeeded - (sm.usage_n ?? sm.succeeded)
      )
      const notRun =
        sm.not_run ?? Math.max(0, sm.total - (sm.attempted ?? sm.total))
      const missingTTFT = input.stressConfig?.stream
        ? Math.max(0, sm.succeeded - sm.ttft_n)
        : 0
      const missingTPOT = input.stressConfig?.stream
        ? Math.max(0, sm.ttft_n - sm.tpot_n)
        : 0
      const slowRows =
        input.stressAssessment?.rows.filter(
          (row) => row.verdict === 'slow' || row.verdict === 'abnormal'
        ) ?? []
      const issueLines = (sm.issues ?? []).map((issue) => {
        const status = issue.status_code || t('No HTTP response')
        return `<li>${issue.count} × ${escapeHtml(String(status))}: ${escapeHtml(issue.message)} (W${issue.worker}/R${issue.round}, ${(issue.elapsed_ms / 1000).toFixed(2)} s)</li>`
      })
      if (sm.other_issue_count) {
        issueLines.push(
          `<li>${escapeHtml(t('{{count}} additional failures are included in the failed total but omitted from the detailed list.', { count: sm.other_issue_count }))}</li>`
        )
      }
      if (sm.failed > 0 && !sm.issues?.length && !sm.other_issue_count) {
        issueLines.push(
          `<li>${escapeHtml(t('Failure details are unavailable for this run.'))}</li>`
        )
      }
      if (missingUsage > 0) {
        issueLines.push(
          `<li>${escapeHtml(t('{{count}} successful requests had no complete token usage; token rates and TPM are unavailable.', { count: missingUsage }))}</li>`
        )
      }
      if (notRun > 0) {
        issueLines.push(
          `<li>${escapeHtml(t('{{count}} planned requests were not started because the run stopped or was cancelled.', { count: notRun }))}</li>`
        )
      }
      if (missingTTFT > 0) {
        issueLines.push(
          `<li>${escapeHtml(t('{{count}} successful stream requests had no observable first output; TTFT is unavailable for them.', { count: missingTTFT }))}</li>`
        )
      }
      if (missingTPOT > 0) {
        issueLines.push(
          `<li>${escapeHtml(t('{{count}} streamed responses lacked enough timed output and token usage to calculate TPOT.', { count: missingTPOT }))}</li>`
        )
      }
      for (const row of slowRows) {
        issueLines.push(
          `<li>${escapeHtml(t(row.label))}: ${escapeHtml(displayMeasured(row, t))} · ${escapeHtml(displayThreshold(row, t))}</li>`
        )
      }
      if (issueLines.length === 0) {
        issueLines.push(
          `<li>${escapeHtml(t('No request errors or threshold violations were found in this run.'))}</li>`
        )
      }
      problemsHtml = `<div class="summary-text"><strong>${escapeHtml(t('Problems found'))}</strong><ul>${issueLines.join('')}</ul></div>`
    }

    const tableHtml = input.stressAssessment
      ? htmlBenchmarkTable(input.stressAssessment, t)
      : ''
    const summaryHtml = input.summaries.stress
      ? `<div class="summary-text">${escapeHtml(input.summaries.stress)}</div>`
      : ''

    stressSection = `<div class="section page-break-avoid">
<div class="section-header">
  <div class="section-title">二、${escapeHtml(t('Concurrency and stress test'))}</div>
  <div class="verdict-badge ${overallCls}">${escapeHtml(badgeText)}</div>
</div>
${configBar}
${statsGrid}
${problemsHtml}
${tableHtml}
${summaryHtml}
</div>`
  }

  // Section 3: Cache test HTML
  let cacheSection = ''
  if (input.cacheAssessment || input.cacheMetrics || input.summaries.cache) {
    const overall = input.cacheAssessment?.overall ?? 'ok'
    const overallCls = verdictHtmlClass(overall)
    const badgeText = `${t('Prompt cache assessment')}：${t(overallLabel(overall))}`

    let configBar = ''
    if (input.cacheMetrics) {
      const cm = input.cacheMetrics
      const modeLabel =
        cm.mode === 'cumulative' ? t('Cumulative chat') : t('Static prefix')
      const freqText =
        cm.hit_count !== undefined
          ? `${cm.hit_count}/${cm.rounds} (${formatPercent(cm.rounds > 0 ? (cm.hit_count ?? 0) / cm.rounds : 0)})`
          : '-'
      const depthText =
        cm.avg_depth_rate !== undefined ? formatPercent(cm.avg_depth_rate) : '-'
      configBar = `<div class="config-bar">
<div class="config-item">${escapeHtml(t('Cache mode'))}：<span>${escapeHtml(modeLabel)}</span></div>
<div class="config-item">${escapeHtml(t('Probe rounds'))}：<span>${cm.rounds}</span></div>
<div class="config-item">${escapeHtml(t('Hit frequency'))}：<span>${escapeHtml(freqText)}</span></div>
<div class="config-item">${escapeHtml(t('Hit depth'))}：<span>${escapeHtml(depthText)}</span></div>
</div>`
    }

    const cacheChecksHtml =
      input.cacheChecks.length > 0
        ? `<div class="check-group-title">${escapeHtml(t('Cache probe details'))}</div>${renderCheckListHtml(input.cacheChecks, input, t)}`
        : ''

    const tableHtml = input.cacheAssessment
      ? htmlBenchmarkTable(input.cacheAssessment, t)
      : ''
    const summaryHtml = input.summaries.cache
      ? `<div class="summary-text">${escapeHtml(input.summaries.cache)}</div>`
      : ''

    cacheSection = `<div class="section page-break-avoid">
<div class="section-header">
  <div class="section-title">三、${escapeHtml(t('Prompt cache test'))}</div>
  <div class="verdict-badge ${overallCls}">${escapeHtml(badgeText)}</div>
</div>
${cacheChecksHtml}
${configBar}
${tableHtml}
${summaryHtml}
</div>`
  }

  return `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<title>${escapeHtml(t('Supplier Test Report'))}-${stampFileName()}</title>
<style>
@page {
  size: A4;
  margin: 0;
}
@media print {
  html, body {
    margin: 0 !important;
    padding: 12mm 16mm !important;
    -webkit-print-color-adjust: exact;
    print-color-adjust: exact;
  }
  .page-break-avoid {
    break-inside: avoid;
    page-break-inside: avoid;
  }
}
* {
  box-sizing: border-box;
}
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
  color: #1e293b;
  line-height: 1.5;
  background: #fff;
  margin: 0 auto;
  padding: 12mm 16mm;
  max-width: 960px;
  font-size: 11.5px;
}

/* Header */
.report-header {
  border-bottom: 2px solid #2563eb;
  padding-bottom: 8px;
  margin-bottom: 10px;
}
.report-title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}
.report-title {
  font-size: 19px;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}
.report-subtitle {
  font-size: 11px;
  color: #64748b;
  margin-top: 2px;
}
.report-badge-top {
  display: inline-flex;
  align-items: center;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  color: #1d4ed8;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
}

/* Meta Grid */
.meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 5px;
  padding: 6px 12px;
  margin-bottom: 12px;
}
.meta-item {
  font-size: 11px;
  display: flex;
}
.meta-label {
  color: #64748b;
  font-weight: 500;
  width: 80px;
  flex-shrink: 0;
}
.meta-value {
  color: #0f172a;
  font-weight: 600;
  word-break: break-all;
}

/* Section */
.section {
  margin-bottom: 14px;
}
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
  padding-bottom: 4px;
  border-bottom: 1px solid #e2e8f0;
}
.section-title {
  font-size: 13.5px;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}
.section-title::before {
  content: "";
  display: inline-block;
  width: 3.5px;
  height: 13px;
  background: #2563eb;
  border-radius: 2px;
}

/* Verdict Badges */
.verdict-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}
.verdict-ok {
  background: #dcfce7;
  color: #15803d;
  border: 1px solid #86efac;
}
.verdict-slow {
  background: #fef9c3;
  color: #a16207;
  border: 1px solid #fde047;
}
.verdict-abnormal {
  background: #fee2e2;
  color: #b91c1c;
  border: 1px solid #fca5a5;
}
.verdict-na {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #cbd5e1;
}

/* Detailed Check List */
.check-group-title {
  font-weight: 600;
  font-size: 11.5px;
  color: #334155;
  margin: 6px 0 3px;
}
.check-list {
  list-style: none;
  padding: 0;
  margin: 0 0 6px 0;
  display: flex;
  flex-direction: column;
  gap: 2.5px;
}
.check-list-item {
  font-size: 11px;
  line-height: 1.4;
  color: #334155;
  display: flex;
  align-items: baseline;
  gap: 5px;
}
.check-bullet {
  color: #3b82f6;
  font-size: 12px;
  line-height: 1;
}
.check-name {
  font-weight: 600;
  color: #1e293b;
  flex-shrink: 0;
}
.check-status-pass {
  color: #15803d;
  font-weight: 600;
  flex-shrink: 0;
}
.check-status-skip {
  color: #64748b;
  font-weight: 600;
  flex-shrink: 0;
}
.check-status-fail {
  color: #dc2626;
  font-weight: 600;
  flex-shrink: 0;
}
.check-detail {
  color: #475569;
}

/* Config Grid */
.config-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 6px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 5px;
  padding: 5px 8px;
  margin-bottom: 5px;
  font-size: 11px;
}
.config-item {
  color: #475569;
}
.config-item span {
  color: #0f172a;
  font-weight: 600;
}

/* Stat Cards */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 5px;
  margin-bottom: 5px;
}
.stat-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 5px;
  padding: 4px 6px;
  text-align: center;
}
.stat-val {
  font-size: 12.5px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.2;
}
.stat-lbl {
  font-size: 10px;
  color: #64748b;
  margin-top: 1px;
}

/* Table */
table.benchmark-table {
  width: 100%;
  border-collapse: collapse;
  margin: 5px 0;
  font-size: 11px;
}
table.benchmark-table th, table.benchmark-table td {
  border: 1px solid #e2e8f0;
  padding: 4px 7px;
  text-align: left;
}
table.benchmark-table th {
  background: #f8fafc;
  font-weight: 600;
  color: #475569;
}
table.benchmark-table td.cell-measured {
  font-weight: 600;
  color: #0f172a;
}
table.benchmark-table td.cell-threshold {
  color: #64748b;
}

.summary-text {
  font-size: 10.5px;
  color: #64748b;
  background: #f8fafc;
  border-left: 3px solid #cbd5e1;
  padding: 4px 8px;
  margin-top: 5px;
  border-radius: 0 4px 4px 0;
}
.summary-text strong {
  color: #334155;
}
.summary-text ul {
  margin: 3px 0 0;
  padding-left: 17px;
}
.summary-text li {
  margin: 2px 0;
  overflow-wrap: anywhere;
}
thead {
  display: table-header-group;
}
tr {
  break-inside: avoid;
  page-break-inside: avoid;
}
.report-footer {
  border-top: 1px solid #e2e8f0;
  color: #94a3b8;
  font-size: 9px;
  margin-top: 16px;
  padding-top: 5px;
  text-align: center;
}
</style>
</head>
<body>

<div class="report-header">
  <div class="report-title-row">
    <div>
      <h1 class="report-title">${escapeHtml(t('Supplier Test Report'))}</h1>
      <div class="report-subtitle">Supplier Benchmark &amp; SLA Compliance Test Report</div>
    </div>
    <div class="report-badge-top">${escapeHtml(t('Commercial evaluation'))}</div>
  </div>
</div>

<div class="meta-grid">
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Model'))}：</span><span class="meta-value">${escapeHtml(input.model || '-')}</span></div>
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Base URL'))}：</span><span class="meta-value">${escapeHtml(input.baseUrl || '-')}</span></div>
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Judgment standard'))}：</span><span class="meta-value">${escapeHtml(input.standardLabel)}</span></div>
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Time'))}：</span><span class="meta-value">${escapeHtml(new Date().toLocaleString())}</span></div>
</div>

<!-- 模块一：接口连通与协议兼容 -->
<div class="section page-break-avoid">
  <div class="section-header">
    <div class="section-title">一、${escapeHtml(t('Connectivity and protocol'))}</div>
    <div class="verdict-badge ${basicBadgeClass}">${escapeHtml(t(basicBadge))}</div>
  </div>
  ${shallowChecks.length > 0 ? `<div class="check-group-title">1. ${escapeHtml(t('Basic connectivity checks'))}</div>${renderCheckListHtml(shallowChecks, input, t)}` : ''}
  ${protocolChecks.length > 0 ? `<div class="check-group-title">2. ${escapeHtml(t('Advanced protocol capabilities'))}</div>${renderCheckListHtml(protocolChecks, input, t)}` : ''}
  ${input.summaries.basic ? `<div class="summary-text">${escapeHtml(input.summaries.basic)}</div>` : ''}
</div>

${stressSection}
${cacheSection}

${input.errorMessage ? `<div class="section page-break-avoid"><h2>${escapeHtml(t('Error'))}</h2><p style="color:#dc2626">${escapeHtml(input.errorMessage)}</p></div>` : ''}

<div class="report-footer">${escapeHtml(t('Generated by Supplier Test'))} · ${escapeHtml(new Date().toLocaleString())}</div>

</body>
</html>
`
}

export type VideoReportInput = {
  baseUrl: string
  model: string
  endpointUrl?: string
  taskId?: string
  status?: string
  elapsedMs: number
  failReason?: string
  videoUrl?: string
  videoConfig: {
    prompt: string
    hasImage: boolean
    uploadMode: string
    imageUrl?: string
    role?: string
    hasLastFrame: boolean
    lastFrameMode?: string
    lastFrameUrl?: string
    resolution?: string
    ratio?: string
    duration?: number
    watermark?: boolean
    seed?: string
    generateAudio?: boolean
    returnLastFrame?: boolean
    customJson?: string
    customPath?: string
  }
  videoChecks: CheckResult[]
  t: (key: string, options?: Record<string, string | number>) => string
}

export function buildVideoMarkdownReport(input: VideoReportInput): string {
  const t = (key: string, options?: Record<string, string | number>) =>
    translate(input.t, key, options)

  const lines = [
    `# ${t('Doubao Video Generation Test Report')}`,
    '',
    `- ${t('Time')}: ${new Date().toLocaleString()}`,
    `- ${t('Base URL')}: ${input.baseUrl || '-'}`,
    `- ${t('Endpoint URL')}: POST ${input.endpointUrl || '-'}`,
    `- ${t('Model')}: ${input.model || '-'}`,
    `- ${t('Task ID')}: ${input.taskId || '-'}`,
    `- ${t('Status')}: ${input.status || t('Unknown')}`,
    `- ${t('Duration')}: ${(input.elapsedMs / 1000).toFixed(1)}s`,
    '',
    `## 1. ${t('Video Generation Parameters')}`,
    `### 1.1 ${t('Prompt')}`,
    input.videoConfig.prompt || '-',
    '',
  ]

  if (input.videoConfig.hasImage) {
    const roleText =
      input.videoConfig.role === 'first_frame'
        ? t('First Frame')
        : t('Reference Image')
    const srcText =
      input.videoConfig.uploadMode === 'url'
        ? input.videoConfig.imageUrl || '-'
        : t('Local File (Base64)')
    lines.push(`- ${t('Image Input')}: ${roleText} (${srcText})`)
  } else {
    lines.push(`- ${t('Input Mode')}: ${t('Text-to-Video')}`)
  }

  if (input.videoConfig.hasLastFrame) {
    const endSrc =
      input.videoConfig.lastFrameMode === 'url'
        ? input.videoConfig.lastFrameUrl || '-'
        : t('Local File (Base64)')
    lines.push(`- ${t('End Frame')}: ${endSrc}`)
  }

  if (input.videoConfig.resolution) {
    lines.push(
      `- ${t('Resolution (resolution)')}: ${input.videoConfig.resolution}`
    )
  }
  if (input.videoConfig.ratio) {
    lines.push(`- ${t('Aspect Ratio (ratio)')}: ${input.videoConfig.ratio}`)
  }
  if (input.videoConfig.duration) {
    lines.push(
      `- ${t('Duration (duration, sec)')}: ${input.videoConfig.duration}s`
    )
  }
  if (input.videoConfig.watermark !== undefined) {
    const wmText = input.videoConfig.watermark ? t('Enabled') : t('Disabled')
    lines.push(`- ${t('Watermark (watermark)')}: ${wmText}`)
  }
  if (input.videoConfig.seed) {
    lines.push(`- ${t('Seed (seed)')}: ${input.videoConfig.seed}`)
  }
  if (input.videoConfig.generateAudio !== undefined) {
    const gaText = input.videoConfig.generateAudio
      ? t('Enabled')
      : t('Disabled')
    lines.push(`- ${t('Generate Audio (generate_audio)')}: ${gaText}`)
  }
  if (input.videoConfig.returnLastFrame !== undefined) {
    const rlfText = input.videoConfig.returnLastFrame
      ? t('Enabled')
      : t('Disabled')
    lines.push(`- ${t('Return Last Frame (return_last_frame)')}: ${rlfText}`)
  }
  if (input.videoConfig.customJson) {
    lines.push(
      `- ${t('Custom Extra Parameters')}: \`${input.videoConfig.customJson}\``
    )
  }

  lines.push('', `## 2. ${t('Execution Pipeline')}`)
  for (const check of input.videoChecks) {
    const detail = check.message ? ` — ${check.message}` : ''
    lines.push(`- ${t(check.title)}: ${t(check.status)}${detail}`)
  }

  lines.push('', `## 3. ${t('Output Result')}`)
  if (input.videoUrl) {
    lines.push(
      `- ${t('Generated Video Output')}: [${t('Open URL in new tab')}](${input.videoUrl})`
    )
  } else if (input.failReason) {
    lines.push(`- ${t('Generation Failed')}: ${input.failReason}`)
  } else {
    lines.push(`- ${t('No output available')}`)
  }

  return `${lines.join('\n')}\n`
}

export function buildVideoHtmlReport(input: VideoReportInput): string {
  const t = (key: string, options?: Record<string, string | number>) =>
    translate(input.t, key, options)
  const statusStr = (input.status || '').toLowerCase()
  const isSucceeded = statusStr === 'succeeded' || statusStr === 'success'
  const isFailed = statusStr === 'failed' || statusStr === 'failure'

  let statusColor = '#2563eb'
  let statusBg = '#eff6ff'
  let statusBorder = '#bfdbfe'
  if (isSucceeded) {
    statusColor = '#15803d'
    statusBg = '#f0fdf4'
    statusBorder = '#bbf7d0'
  } else if (isFailed) {
    statusColor = '#dc2626'
    statusBg = '#fef2f2'
    statusBorder = '#fecaca'
  }

  const checksHtml = input.videoChecks
    .map((check) => {
      let statusCls = 'check-status-fail'
      if (check.status === 'pass') {
        statusCls = 'check-status-pass'
      } else if (check.status === 'skip') {
        statusCls = 'check-status-skip'
      }
      return `<div class="check-item"><span class="${statusCls}">●</span><span class="check-title">${escapeHtml(t(check.title))}：</span><span class="check-status-text">${escapeHtml(t(check.status))}</span>${check.message ? `<span class="check-msg">${escapeHtml(check.message)}</span>` : ''}</div>`
    })
    .join('')

  const params: string[] = []
  if (input.videoConfig.hasImage) {
    const roleText =
      input.videoConfig.role === 'first_frame'
        ? t('First Frame')
        : t('Reference Image')
    const srcText =
      input.videoConfig.uploadMode === 'url'
        ? input.videoConfig.imageUrl || '-'
        : t('Local File (Base64)')
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('Image Input'))}：</span><span class="meta-value">${escapeHtml(roleText)} (${escapeHtml(srcText)})</span></div>`
    )
  } else {
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('Input Mode'))}：</span><span class="meta-value">${escapeHtml(t('Text-to-Video'))}</span></div>`
    )
  }
  if (input.videoConfig.hasLastFrame) {
    const endSrc =
      input.videoConfig.lastFrameMode === 'url'
        ? input.videoConfig.lastFrameUrl || '-'
        : t('Local File (Base64)')
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('End Frame'))}：</span><span class="meta-value">${escapeHtml(endSrc)}</span></div>`
    )
  }
  if (input.videoConfig.resolution) {
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('Resolution (resolution)'))}：</span><span class="meta-value">${escapeHtml(input.videoConfig.resolution)}</span></div>`
    )
  }
  if (input.videoConfig.ratio) {
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('Aspect Ratio (ratio)'))}：</span><span class="meta-value">${escapeHtml(input.videoConfig.ratio)}</span></div>`
    )
  }
  if (input.videoConfig.duration) {
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('Duration (duration, sec)'))}：</span><span class="meta-value">${input.videoConfig.duration}s</span></div>`
    )
  }
  if (input.videoConfig.watermark !== undefined) {
    const wm = input.videoConfig.watermark ? t('Enabled') : t('Disabled')
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('Watermark (watermark)'))}：</span><span class="meta-value">${escapeHtml(wm)}</span></div>`
    )
  }
  if (input.videoConfig.seed) {
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('Seed (seed)'))}：</span><span class="meta-value">${escapeHtml(input.videoConfig.seed)}</span></div>`
    )
  }
  if (input.videoConfig.generateAudio !== undefined) {
    const ga = input.videoConfig.generateAudio ? t('Enabled') : t('Disabled')
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('Generate Audio (generate_audio)'))}：</span><span class="meta-value">${escapeHtml(ga)}</span></div>`
    )
  }
  if (input.videoConfig.returnLastFrame !== undefined) {
    const rlf = input.videoConfig.returnLastFrame ? t('Enabled') : t('Disabled')
    params.push(
      `<div class="meta-item"><span class="meta-label">${escapeHtml(t('Return Last Frame (return_last_frame)'))}：</span><span class="meta-value">${escapeHtml(rlf)}</span></div>`
    )
  }
  if (input.videoConfig.customJson) {
    params.push(
      `<div class="meta-item" style="grid-column: span 2;"><span class="meta-label">${escapeHtml(t('Custom Extra Parameters'))}：</span><span class="meta-value font-mono">${escapeHtml(input.videoConfig.customJson)}</span></div>`
    )
  }

  let outputSectionHtml = `<p style="color:#64748b">${escapeHtml(t('No output available'))}</p>`
  if (input.videoUrl) {
    outputSectionHtml = `
    <div class="output-box">
      <div class="output-title">✓ ${escapeHtml(t('Generated Video Output'))}</div>
      <div class="output-url"><a href="${escapeHtml(input.videoUrl)}" target="_blank" rel="noreferrer">${escapeHtml(input.videoUrl)}</a></div>
    </div>`
  } else if (input.failReason) {
    outputSectionHtml = `
    <div class="error-box">
      <div class="error-title">✕ ${escapeHtml(t('Generation Failed'))}</div>
      <div class="error-msg">${escapeHtml(input.failReason)}</div>
    </div>`
  }

  return `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<title>${escapeHtml(t('Doubao Video Generation Test Report'))}-${stampFileName()}</title>
<style>
@page { size: A4; margin: 0; }
@media print {
  html, body {
    margin: 0 !important;
    padding: 12mm 16mm !important;
    -webkit-print-color-adjust: exact;
    print-color-adjust: exact;
  }
  .page-break-avoid { break-inside: avoid; page-break-inside: avoid; }
}
* { box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
  color: #1e293b;
  line-height: 1.5;
  background: #fff;
  margin: 0 auto;
  padding: 12mm 16mm;
  max-width: 960px;
  font-size: 11.5px;
}
.report-header { border-bottom: 2px solid #7c3aed; padding-bottom: 8px; margin-bottom: 12px; }
.report-title-row { display: flex; justify-content: space-between; align-items: flex-end; }
.report-title { font-size: 19px; font-weight: 700; color: #0f172a; margin: 0; }
.report-subtitle { font-size: 11px; color: #64748b; margin-top: 2px; }
.report-badge-top { display: inline-flex; align-items: center; background: ${statusBg}; border: 1px solid ${statusBorder}; color: ${statusColor}; font-size: 11px; font-weight: 700; padding: 3px 10px; border-radius: 4px; text-transform: uppercase; }
.meta-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 4px 16px; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 6px; padding: 8px 12px; margin-bottom: 12px; }
.meta-item { display: flex; align-items: baseline; font-size: 11px; }
.meta-label { color: #64748b; white-space: nowrap; flex-shrink: 0; }
.meta-value { color: #0f172a; font-weight: 500; word-break: break-all; }
.section { margin-bottom: 12px; border: 1px solid #e2e8f0; border-radius: 6px; padding: 10px 14px; background: #fff; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; border-bottom: 1px solid #f1f5f9; padding-bottom: 4px; }
.section-title { font-size: 13px; font-weight: 700; color: #0f172a; }
.prompt-box { font-size: 11px; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 4px; padding: 8px 10px; white-space: pre-wrap; word-break: break-word; color: #334155; margin-bottom: 8px; }
.check-item { display: flex; align-items: baseline; gap: 6px; font-size: 11px; padding: 3px 0; border-bottom: 1px dashed #f1f5f9; }
.check-item:last-child { border-bottom: none; }
.check-status-pass { color: #16a34a; font-size: 9px; flex-shrink: 0; }
.check-status-fail { color: #dc2626; font-size: 9px; flex-shrink: 0; }
.check-status-skip { color: #94a3b8; font-size: 9px; flex-shrink: 0; }
.check-title { font-weight: 600; color: #1e293b; flex-shrink: 0; }
.check-status-text { font-weight: 600; text-transform: uppercase; font-size: 10.5px; }
.check-msg { color: #64748b; font-size: 10.5px; word-break: break-all; }
.output-box { background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 6px; padding: 10px 14px; margin-top: 8px; }
.output-title { font-weight: 600; color: #15803d; font-size: 12px; margin-bottom: 4px; }
.output-url { font-family: monospace; font-size: 11px; word-break: break-all; color: #166534; }
.output-url a { color: #166534; text-decoration: underline; }
.error-box { background: #fef2f2; border: 1px solid #fecaca; border-radius: 6px; padding: 10px 14px; margin-top: 8px; }
.error-title { font-weight: 600; color: #b91c1c; font-size: 12px; margin-bottom: 4px; }
.error-msg { font-size: 11px; color: #991b1b; word-break: break-all; }
</style>
</head>
<body>

<div class="report-header">
  <div class="report-title-row">
    <div>
      <h1 class="report-title">${escapeHtml(t('Doubao Video Generation Test Report'))}</h1>
      <div class="report-subtitle">Volcano Ark Seedance Video Model Verification</div>
    </div>
    <div class="report-badge-top">${escapeHtml(input.status || t('Unknown'))}</div>
  </div>
</div>

<div class="meta-grid">
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Model'))}：</span><span class="meta-value">${escapeHtml(input.model || '-')}</span></div>
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Base URL'))}：</span><span class="meta-value">${escapeHtml(input.baseUrl || '-')}</span></div>
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Endpoint URL'))}：</span><span class="meta-value">POST ${escapeHtml(input.endpointUrl || '-')}</span></div>
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Task ID'))}：</span><span class="meta-value">${escapeHtml(input.taskId || '-')}</span></div>
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Duration'))}：</span><span class="meta-value">${(input.elapsedMs / 1000).toFixed(1)} s</span></div>
  <div class="meta-item"><span class="meta-label">${escapeHtml(t('Time'))}：</span><span class="meta-value">${escapeHtml(new Date().toLocaleString())}</span></div>
</div>

<div class="section page-break-avoid">
  <div class="section-header">
    <div class="section-title">一、${escapeHtml(t('Video Generation Parameters'))}</div>
  </div>
  <div class="prompt-box"><strong>${escapeHtml(t('Prompt'))}：</strong>${escapeHtml(input.videoConfig.prompt || '-')}</div>
  <div class="meta-grid" style="margin-bottom:0; background:#fff; border:none; padding:0;">
    ${params.join('')}
  </div>
</div>

<div class="section page-break-avoid">
  <div class="section-header">
    <div class="section-title">二、${escapeHtml(t('Execution Pipeline'))}</div>
  </div>
  <div class="checks-list">
    ${checksHtml}
  </div>
</div>

<div class="section page-break-avoid">
  <div class="section-header">
    <div class="section-title">三、${escapeHtml(t('Output Result'))}</div>
  </div>
  ${outputSectionHtml}
</div>

</body>
</html>`
}

export function exportVideoPdfReport(input: VideoReportInput) {
  const html = buildVideoHtmlReport(input)
  const iframe = document.createElement('iframe')
  iframe.style.position = 'fixed'
  iframe.style.right = '0'
  iframe.style.bottom = '0'
  iframe.style.width = '0'
  iframe.style.height = '0'
  iframe.style.border = '0'
  iframe.title = 'doubao-video-report-print'
  document.body.appendChild(iframe)

  const doc = iframe.contentWindow?.document
  if (!doc) {
    document.body.removeChild(iframe)
    return
  }

  doc.open()
  doc.write(html)
  doc.close()

  printReportFrame(iframe, doc)
}
