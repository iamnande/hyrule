package models

type BillingPlan string

const (
	BillingPlanConsumption BillingPlan = "consumption"
	BillingPlanContract    BillingPlan = "contract"
)

func (plan BillingPlan) String() string {
	return string(plan)
}
