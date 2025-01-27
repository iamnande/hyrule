package models

type BillingPlan string

const (
	BillingPlanConsumption BillingPlan = "consumption"
	BillingPlanContract    BillingPlan = "contract"
)
