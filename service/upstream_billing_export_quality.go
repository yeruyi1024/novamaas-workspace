package service

import (
	"fmt"
	"strings"
)

// Unknown amounts have no numeric value. Partially known amounts remain usable
// subtotals, with their scope carried alongside the amount in every export.
func upstreamBillingExportCostScope(row UpstreamBillingRow) string {
	if row.MissingCostRecords == 0 {
		return "成本完整"
	}
	if row.Count <= row.MissingCostRecords {
		return "成本未知"
	}
	return "已知成本小计（不含未知金额）"
}

func upstreamBillingExportAmount(value string, row UpstreamBillingRow) string {
	if row.MissingCostRecords > 0 && row.Count <= row.MissingCostRecords {
		return "未知"
	}
	return value
}

func upstreamBillingExportNetLabel(snapshot *UpstreamBillingSnapshot) string {
	row := snapshot.Total
	if !upstreamBillingExportHasEligibleChannel(snapshot) || (row.MissingCostRecords > 0 && row.Count <= row.MissingCostRecords) {
		return "净成本未知"
	}
	if row.MissingCostRecords > 0 || snapshot.Coverage.MissingChannels > 0 {
		return "已知净成本小计"
	}
	return "净成本"
}

func upstreamBillingExportHasEligibleChannel(snapshot *UpstreamBillingSnapshot) bool {
	for _, channel := range snapshot.Group.Channels {
		if channel.Eligible && !channel.Missing {
			return true
		}
	}
	return false
}

func upstreamBillingExportTotalAmount(value string, snapshot *UpstreamBillingSnapshot) string {
	if !upstreamBillingExportHasEligibleChannel(snapshot) {
		return "未知"
	}
	return upstreamBillingExportAmount(value, snapshot.Total)
}

func upstreamBillingExportUsageStatus(row UpstreamBillingRow) string {
	parts := make([]string, 0, 3)
	if row.MissingUsageRecords > 0 {
		parts = append(parts, fmt.Sprintf("用量未知 %d 条", row.MissingUsageRecords))
	}
	if row.EstimatedUsageRecords > 0 {
		parts = append(parts, fmt.Sprintf("估算用量 %d 条", row.EstimatedUsageRecords))
	}
	if row.PendingTasks > 0 {
		parts = append(parts, fmt.Sprintf("待结算 %d 条", row.PendingTasks))
	}
	if len(parts) == 0 {
		return "用量证据完整"
	}
	return strings.Join(parts, " / ")
}

func upstreamBillingExportQualityWarning(snapshot *UpstreamBillingSnapshot) string {
	coverage := snapshot.Coverage
	if coverage.Complete {
		return "数据质量检查：未发现缺口。"
	}
	message := fmt.Sprintf("数据质量告警：缺失渠道 %d，缺少成本 %d，用量未知 %d，估算用量 %d，待结算 %d。", coverage.MissingChannels, coverage.MissingCostRecords, coverage.MissingUsageRecords, coverage.EstimatedUsageRecords, coverage.PendingTasks)
	if !upstreamBillingExportHasEligibleChannel(snapshot) {
		message += "无可纳入成本统计的渠道。"
	}
	return message + "用于核对，缺口待核实。"
}
