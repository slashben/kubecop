package rule

// List of all rules descriptions.
var ruleDescriptions []RuleDesciptor = []RuleDesciptor{
	R0001ExecWhitelistedRuleDescriptor,
}

func CreateRulesByTags(tags []string) []Rule {
	var rules []Rule
	for _, rule := range ruleDescriptions {
		if rule.HasTags(tags) {
			rules = append(rules, rule.RuleCreationFunc())
		}
	}
	return rules
}

func CreateRulesByNames(names []string) []Rule {
	var rules []Rule
	for _, rule := range ruleDescriptions {
		for _, name := range names {
			if rule.Name == name {
				rules = append(rules, rule.RuleCreationFunc())
			}
		}
	}
	return rules
}
