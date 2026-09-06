#!/usr/bin/env bash
set -euo pipefail

readme=README.md
start_marker='<!-- novamaas-pr-ledger:start -->'
end_marker='<!-- novamaas-pr-ledger:end -->'

if [[ ! -f "$readme" ]]; then
  echo "$readme is missing." >&2
  exit 1
fi

start_count=$(grep -Fxc "$start_marker" "$readme" || true)
end_count=$(grep -Fxc "$end_marker" "$readme" || true)
if [[ "$start_count" != 1 || "$end_count" != 1 ]]; then
  echo "README PR ledger markers must each appear exactly once." >&2
  exit 1
fi

start_line=$(awk -v marker="$start_marker" '$0 == marker { print NR }' "$readme")
end_line=$(awk -v marker="$end_marker" '$0 == marker { print NR }' "$readme")
if ((start_line >= end_line)); then
  echo "README PR ledger markers are out of order." >&2
  exit 1
fi

if ! sed -n "${start_line},${end_line}p" "$readme" | grep -Fq '| PR / 工作项 | 日期 | 类型 | 领域 | 关键变化 | 与上游关系 | 状态 |'; then
  echo "README PR ledger table header is missing." >&2
  exit 1
fi

pr_number=${README_LEDGER_PR_NUMBER:-}
if [[ -z "$pr_number" ]]; then
  exit 0
fi
if [[ ! "$pr_number" =~ ^[1-9][0-9]*$ ]]; then
  echo "Invalid README_LEDGER_PR_NUMBER: $pr_number" >&2
  exit 1
fi

repository=${README_LEDGER_REPOSITORY:?README_LEDGER_REPOSITORY is required for PR validation}
entry_prefix="| [#${pr_number}](https://github.com/${repository}/pull/${pr_number}) |"
entry_count=$(sed -n "${start_line},${end_line}p" "$readme" | grep -Fc "$entry_prefix" || true)
if [[ "$entry_count" != 1 ]]; then
  echo "Add the current PR to the README ledger using: $entry_prefix" >&2
  exit 1
fi
if sed -n "${start_line},${end_line}p" "$readme" | grep -Eq '^\| (TBD|#TBD|待定|待提交)'; then
  echo "Replace provisional README ledger work items with real PR links before merging." >&2
  exit 1
fi

base_sha=${README_LEDGER_BASE_SHA:?README_LEDGER_BASE_SHA is required for PR validation}
head_sha=${README_LEDGER_HEAD_SHA:?README_LEDGER_HEAD_SHA is required for PR validation}
git cat-file -e "${base_sha}^{commit}"
git cat-file -e "${head_sha}^{commit}"
if ! git diff --unified=0 "$base_sha" "$head_sha" -- "$readme" | grep -Fq "+${entry_prefix}"; then
  echo "The current branch must add its own PR row inside the README ledger." >&2
  exit 1
fi
