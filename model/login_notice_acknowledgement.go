package model

// LoginNoticeAcknowledgement records that a dashboard login session saw and
// acknowledged the post-login notice. DeviceFingerprint stores a one-way hash
// produced by the browser; raw device attributes are never persisted.
type LoginNoticeAcknowledgement struct {
	ID                      int64  `json:"id" gorm:"primaryKey"`
	UserID                  int    `json:"user_id" gorm:"index;uniqueIndex:idx_login_notice_ack_user_session,priority:1"`
	SessionID               string `json:"session_id" gorm:"type:varchar(64);uniqueIndex:idx_login_notice_ack_user_session,priority:2"`
	DeviceFingerprint       string `json:"device_fingerprint" gorm:"type:char(64)"`
	IP                      string `json:"ip" gorm:"type:varchar(64)"`
	UserAgent               string `json:"user_agent" gorm:"type:varchar(512)"`
	AnnouncementFingerprint string `json:"announcement_fingerprint" gorm:"type:char(64)"`
	RequiredAcknowledgement bool   `json:"required_acknowledgement"`
	GeneratedToday          int64  `json:"generated_today"`
	GeneratedSevenDays      int64  `json:"generated_seven_days"`
	GeneratedThirtyDays     int64  `json:"generated_thirty_days"`
	ViolationsToday         int64  `json:"violations_today"`
	ViolationsSevenDays     int64  `json:"violations_seven_days"`
	ViolationsThirtyDays    int64  `json:"violations_thirty_days"`
	AcknowledgedAt          int64  `json:"acknowledged_at" gorm:"index"`
}

func HasLoginNoticeAcknowledgement(userID int, sessionID string) (bool, error) {
	var count int64
	err := DB.Model(&LoginNoticeAcknowledgement{}).
		Where("user_id = ? AND session_id = ?", userID, sessionID).
		Count(&count).Error
	return count > 0, err
}

func SaveLoginNoticeAcknowledgement(acknowledgement *LoginNoticeAcknowledgement) error {
	var stored LoginNoticeAcknowledgement
	return DB.Where("user_id = ? AND session_id = ?", acknowledgement.UserID, acknowledgement.SessionID).
		Assign(acknowledgement).
		FirstOrCreate(&stored).Error
}
