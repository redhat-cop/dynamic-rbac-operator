package helpers

import (
	"context"
	"errors"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/redhat-cop/dynamic-rbac-operator/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
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
		if len(*inherit) != 1 {
			// TODO: multi-role inheritance, merging policies, etc.
			return nil, errors.New("this operator only supports one inherited role right now")
		}
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
				rules = append(rules, expandedPolicyRules...)
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
				rules = append(rules, expandedPolicyRules...)
			}
		}
	}

	if deny != nil {
		rules = ApplyDenyRulesToExpandedRuleset(rules, *deny)
	}

	return &rules, nil
}

// EnumeratePolicyRules takes a list of rules with wildcards and returns a list of policy rules with resources explicitly enumerated
func EnumeratePolicyRules(inputRules []rbacv1.PolicyRule, cache *ResourceCache) ([]rbacv1.PolicyRule, error) {
	rules := []rbacv1.PolicyRule{}
	allPossibleRules := cache.AllPolicies
	for _, rule := range inputRules {
		if ruleHasGroupWildcard(&rule) && ruleHasResourceWildcard(&rule) {
			var relevantRules []rbacv1.PolicyRule
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
						var tmpRule rbacv1.PolicyRule
						copier.Copy(&tmpRule, &matchedRule)
						copier.Copy(&tmpRule.Verbs, &rule.Verbs)
						rules = append(rules, tmpRule)
					}
				}
			}
		} else if ruleHasResourceWildcard(&rule) {
			for _, group := range rule.APIGroups {
				for _, matchedRule := range *allPossibleRules {
					if stringInSlice(matchedRule.APIGroups, group) {
						var tmpRule rbacv1.PolicyRule
						copier.Copy(&tmpRule, &matchedRule)
						copier.Copy(&tmpRule.Verbs, &rule.Verbs)
						rules = append(rules, tmpRule)
					}
				}
			}
		} else {
			rules = append(rules, rule)
		}
	}
	return rules, nil
}

// ExpandPolicyRules ensures that multiple resources with the same verbs are not grouped together in the same rule definition (makes it easier to edit individual verbs later)
func ExpandPolicyRules(inputRules []rbacv1.PolicyRule) []rbacv1.PolicyRule {
	rules := []rbacv1.PolicyRule{}
	for _, rule := range inputRules {
		if len(rule.Resources) > 1 {
			for _, resource := range rule.Resources {
				var newVerbs []string
				var newAPIGroups []string
				copier.Copy(&newVerbs, &rule.Verbs)
				copier.Copy(&newAPIGroups, &rule.APIGroups)
				newRule := rbacv1.PolicyRule{
					APIGroups: newAPIGroups,
					Resources: []string{resource},
					Verbs:     newVerbs,
				}
				rules = append(rules, newRule)
			}
		} else {
			var tmpRule rbacv1.PolicyRule
			copier.Copy(&tmpRule, &rule)
			rules = append(rules, tmpRule)
		}
	}
	return rules
}

// APIResourcesToExpandedRules converts an APIResourceList into a list of PolicyRules with all verbs allowed
func APIResourcesToExpandedRules(resourceLists []*metav1.APIResourceList) []rbacv1.PolicyRule {
	policyRules := make([]rbacv1.PolicyRule, 0, 100)

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
			policyRules = append(policyRules, rbacv1.PolicyRule{
				APIGroups: []string{group},
				Resources: []string{resource.Name},
				Verbs:     verbs,
			})
		}
	}

	return policyRules
}

// ApplyDenyRulesToExpandedRuleset takes in an expanded ruleset (see func `ExpandPolicyRules`) and removes anything matching the deny rules
func ApplyDenyRulesToExpandedRuleset(fullRuleSet []rbacv1.PolicyRule, denyRules []rbacv1.PolicyRule) []rbacv1.PolicyRule {
	outputRules := []rbacv1.PolicyRule{}

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
				if !stringInSlice(denyRule.Verbs, "*") {
					newVerbs = subtractStringSlices(rule.Verbs, denyRule.Verbs)
					if len(newVerbs) > 0 {
						var tmpRule rbacv1.PolicyRule
						copier.Copy(&tmpRule, &rule)
						tmpRule.Verbs = newVerbs
						outputRules = append(outputRules, tmpRule)
					}
				}
			} else {
				outputRules = append(outputRules, rule)
			}
		}
	}

	return outputRules
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

func ruleHasGroupWildcard(rule *rbacv1.PolicyRule) bool {
	return stringInSlice(rule.APIGroups, "*")
}

func ruleHasResourceWildcard(rule *rbacv1.PolicyRule) bool {
	return stringInSlice(rule.Resources, "*")
}
