package types

type FreeEnrollmentPlanBody struct {
	Name                   string `json:"name" binding:"required"`
	ReferralSource         string `json:"referral_source" binding:"required"`
	ReferralSourceSpecific string `json:"referral_source_specific"`
}
