package helpers

import (
	"context"
	"errors"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/redhat-cop/dynamic-rbac-operator/api/v1alpha1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RoleType int

const (
	Role RoleType = iota
	ClusterRole
)

// BuildPolicyRules takes an inherited role, an allow list, and a deny list; and processes everything into a list of policy rules
func BuildPolicyRules(client client.Client, cache ResourceCache, roleType RoleType, forNamespace string, inherit *[]v1alpha1.InheritedRole, allow *[]v1.PolicyRule, deny *[]v1.PolicyRule) (*[]v1.PolicyRule, error) {
	rules := []v1.PolicyRule{}

	if inherit != nil {
		for _, roleToInherit := range *inherit {
			switch roleToInherit.Kind {
			case "ClusterRole":
				inheritedClusterRole := &v1.ClusterRole{}
				clusterRoleNamespacedName := types.NamespacedName{Name: roleToInherit.Name}
				err := client.Get(context.TODO(), clusterRoleNamespacedName, inheritedClusterRole)
				if err != nil {
					return nil, err
				}
				// nonResourceURLs do not make sense to move from a ClusterRole to a Role
				var enumeratedPolicyRules []v1.PolicyRule
				if roleType == Role {
					enumeratedPolicyRules, err = EnumeratePolicyRules(StripNonResourceURLs(inheritedClusterRole.Rules), &cache)
				} else {
					enumeratedPolicyRules, err = EnumeratePolicyRules(inheritedClusterRole.Rules, &cache)
				}
				expandedPolicyRules := ExpandPolicyRules(enumeratedPolicyRules)
				rules = MergeExpandedPolicyRules(rules, expandedPolicyRules)
			case "Role":
				if roleType == ClusterRole && roleToInherit.Namespace == "" {
					return nil, errors.New("a Cluster Role cannot inherit from a Role without a namespace specified")
				}
				useNamespace := forNamespace
				if roleToInherit.Namespace != "" {
					useNamespace = roleToInherit.Namespace
				}
				inheritedRole := &v1.Role{}
				roleNamespacedName := types.NamespacedName{Name: roleToInherit.Name, Namespace: useNamespace}
				err := client.Get(context.TODO(), roleNamespacedName, inheritedRole)
				if err != nil {
					return nil, err
				}
				enumeratedPolicyRules, err := EnumeratePolicyRules(inheritedRole.Rules, &cache)
				expandedPolicyRules := ExpandPolicyRules(enumeratedPolicyRules)
				rules = MergeExpandedPolicyRules(rules, expandedPolicyRules)
			}
		}
	}

	if deny != nil {
		rules = ApplyDenyRulesToExpandedRuleset(rules, *deny)
	}

	if allow != nil {
		allowRules, err := EnumeratePolicyRules(*allow)
		if err != nil {
			return nil, err
		}
		rules = MergeExpandedPolicyRules(rules, ExpandPolicyRules(allowRules))
	}

	return &rules, nil
}

// EnumeratePolicyRules takes a list of rules with wildcards and returns a list of policy rules with resources explicitly enumerated
func EnumeratePolicyRules(inputRules []v1.PolicyRule, cache *ResourceCache) ([]v1.PolicyRule, error) {
	rules := []v1.PolicyRule{}
	allPossibleRules := cache.AllPolicies
	for _, rule := range inputRules {
		if ruleHasGroupWildcard(&rule) && ruleHasResourceWildcard(&rule) {
			var relevantRules []v1.PolicyRule
			copier.Copy(&relevantRules, allPossibleRules)
			if !stringInSlice(rule.Verbs, "*") {
				for index := range relevantRules {
					relevantRules[index].Verbs = rule.Verbs
				}
			}
			rules = append(rules, relevantRules...)
		} else if ruleHasGroupWildcard(&rule) {
			for _, resource := range rule.Resources {
				for _, matchedRule := range *allPossibleRules {
					if stringInSlice(matchedRule.Resources, resource) {
						var tmpRule v1.PolicyRule
						copier.Copy(&tmpRule, &matchedRule)
						if !stringInSlice(rule.Verbs, "*") {
							copier.Copy(&tmpRule.Verbs, &rule.Verbs)
						} else {
							copier.Copy(&tmpRule.Verbs, &matchedRule.Verbs)
						}
						rules = append(rules, tmpRule)
					}
				}
			}
		} else if ruleHasResourceWildcard(&rule) {
			for _, group := range rule.APIGroups {
				for _, matchedRule := range *allPossibleRules {
					if stringInSlice(matchedRule.APIGroups, group) {
						var tmpRule v1.PolicyRule
						copier.Copy(&tmpRule, &matchedRule)
						if !stringInSlice(rule.Verbs, "*") {
							copier.Copy(&tmpRule.Verbs, &rule.Verbs)
						} else {
							copier.Copy(&tmpRule.Verbs, &matchedRule.Verbs)
						}
						rules = append(rules, tmpRule)
					}
				}
			}
		} else {
			for _, group := range rule.APIGroups {
				if group == "v1" {
					group = ""
				}
				for _, resource := range rule.Resources {
					for _, matchedRule := range *allPossibleRules {
						if stringInSlice(matchedRule.APIGroups, group) && stringInSlice(matchedRule.Resources, resource) {
							var tmpRule v1.PolicyRule
							copier.Copy(&tmpRule, &matchedRule)
							if !stringInSlice(rule.Verbs, "*") {
								copier.Copy(&tmpRule.Verbs, &rule.Verbs)
							} else {
								copier.Copy(&tmpRule.Verbs, &matchedRule.Verbs)
							}
							rules = append(rules, tmpRule)
						}
					}
				}
			}
		}
	}
	return rules, nil
}

// ExpandPolicyRules ensures that multiple resources with the same verbs are not grouped together in the same rule definition (makes it easier to edit individual verbs later)
func ExpandPolicyRules(inputRules []v1.PolicyRule) []v1.PolicyRule {
	rules := []v1.PolicyRule{}
	for _, rule := range inputRules {
		if len(rule.Resources) > 1 {
			for _, resource := range rule.Resources {
				var newVerbs []string
				var newAPIGroups []string
				copier.Copy(&newVerbs, &rule.Verbs)
				copier.Copy(&newAPIGroups, &rule.APIGroups)
				newRule := v1.PolicyRule{
					APIGroups: newAPIGroups,
					Resources: []string{resource},
					Verbs:     newVerbs,
				}
				rules = append(rules, newRule)
			}
		} else {
			var tmpRule v1.PolicyRule
			copier.Copy(&tmpRule, &rule)
			rules = append(rules, tmpRule)
		}
	}
	return rules
}

// MergeExpandedPolicyRules takes two expanded rulesets (see func `ExpandPolicyRules`) and returns one merged expanded ruleset
func MergeExpandedPolicyRules(first []v1.PolicyRule, second []v1.PolicyRule) []v1.PolicyRule {
	return irToPolicyList(unionIRs(policyListToIR(first), policyListToIR(second)))
}

// APIResourcesToExpandedRules converts an APIResourceList into a list of PolicyRules with all verbs allowed
func APIResourcesToExpandedRules(resourceLists []*metav1.APIResourceList) []v1.PolicyRule {
	outputIR := make(policyListIR)

	for _, resourceList := range resourceLists {
		for _, resource := range resourceList.APIResources {
			group := strings.Split(resourceList.GroupVersion, "/")[0]
			if group == "v1" {
				group = "" // an extremely cool thing to have to do...
			}
			verbs := []string{"*"}
			if len(resource.Verbs) > 0 {
				copier.Copy(&verbs, &resource.Verbs)
			}
			currentPolicyKey := expandedPolicyKey{
				APIGroup: group,
				Resource: resource.Name,
			}
			if _, ok := outputIR[currentPolicyKey]; ok {
				// We do this so that we don't generate multiple policy rules for a resource that exists as multiple versions - i.e. v1alpha1, v1beta1, etc.
				outputIR[currentPolicyKey] = appendSet(outputIR[currentPolicyKey], verbs...)
			} else {
				outputIR[currentPolicyKey] = verbs
			}
		}
	}

	return irToPolicyList(outputIR)
}

// ApplyDenyRulesToExpandedRuleset takes in an expanded ruleset (see func `ExpandPolicyRules`) and removes anything matching the deny rules
func ApplyDenyRulesToExpandedRuleset(fullRuleSet []v1.PolicyRule, denyRules []v1.PolicyRule) []v1.PolicyRule {
	outputIR := policyListToIR(fullRuleSet)

	for _, rule := range fullRuleSet {
		for _, denyRule := range denyRules {
			var newVerbs []string
			denyRuleApplies := false
			if ruleHasGroupWildcard(&denyRule) && ruleHasResourceWildcard(&denyRule) {
				denyRuleApplies = true
			} else if ruleHasGroupWildcard(&denyRule) && slicesIntersect(denyRule.Resources, rule.Resources) {
				denyRuleApplies = true
			} else if ruleHasResourceWildcard(&denyRule) && slicesIntersect(denyRule.APIGroups, rule.APIGroups) {
				denyRuleApplies = true
			} else if slicesIntersect(denyRule.APIGroups, rule.APIGroups) && slicesIntersect(denyRule.Resources, rule.Resources) {
				denyRuleApplies = true
			}
			if denyRuleApplies {
				currentPolicyKey := expandedPolicyKey{
					APIGroup: rule.APIGroups[0],
					Resource: rule.Resources[0],
				}
				if !stringInSlice(denyRule.Verbs, "*") {
					newVerbs = subtractStringSlices(rule.Verbs, denyRule.Verbs)
					if len(newVerbs) > 0 {
						var tmpRule v1.PolicyRule
						copier.Copy(&tmpRule, &rule)
						tmpRule.Verbs = newVerbs
						outputIR[currentPolicyKey] = newVerbs
					} else {
						delete(outputIR, currentPolicyKey)
					}
				} else {
					delete(outputIR, currentPolicyKey)
				}
			}
		}
	}

	return irToPolicyList(outputIR)
}

// StripNonResourceURLs takes a list of PolicyRules that may specify NonResourceURLs and returns the same list without any NonResourceURLs
func StripNonResourceURLs(rules []v1.PolicyRule) []v1.PolicyRule {
	var ln int
	for _, rule := range rules {
		if rule.NonResourceURLs != nil {
			continue
		}
		rules[ln] = rule
		ln++
	}
	return rules[:ln]
}

func ruleHasGroupWildcard(rule *v1.PolicyRule) bool {
	return stringInSlice(rule.APIGroups, "*")
}

func ruleHasResourceWildcard(rule *v1.PolicyRule) bool {
	return stringInSlice(rule.Resources, "*")
}
