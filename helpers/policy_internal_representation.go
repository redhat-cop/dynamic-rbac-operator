package helpers

import (
	v1 "k8s.io/api/rbac/v1"
)

type expandedPolicyKey struct {
	APIGroup        string
	Resource        string
	ResourceNames   string
	NonResourceURLs string
}

type policyValue []string

// PolicyListIR is an internal representation of a list of PolicyRules which is easier to patch and manipulate
type policyListIR map[expandedPolicyKey]policyValue

func policyListToIR(input []v1.PolicyRule) policyListIR {
	outputMap := make(policyListIR)
	for _, rule := range input {
		currentPolicyKey := expandedPolicyKey{
			APIGroup: rule.APIGroups[0],
			Resource: rule.Resources[0],
		}
		if _, ok := outputMap[currentPolicyKey]; ok {
			outputMap[currentPolicyKey] = appendSet(outputMap[currentPolicyKey], rule.Verbs...)
		} else {
			outputMap[currentPolicyKey] = rule.Verbs
		}
	}
	return outputMap
}

func irToPolicyList(input policyListIR) []v1.PolicyRule {
	output := make([]v1.PolicyRule, 0, 100)
	for key, verbs := range input {
		output = append(output, v1.PolicyRule{
			APIGroups: []string{key.APIGroup},
			Resources: []string{key.Resource},
			Verbs:     verbs,
		})
	}
	return output
}

func unionIRs(m1, m2 policyListIR) policyListIR {
	for ia, va := range m1 {
		if it, ok := m2[ia]; ok {
			va = appendSet(va, it...)
		}
		m2[ia] = va

	}
	return m2
}
