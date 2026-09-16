package com.familytimenet.familyguard.core.filter

import com.familytimenet.familyguard.core.model.FamilyPolicy

/** Combines domain and IPv4 policy evaluation without inspecting application content. */
class FamilyGuardFilter(private val policyProvider: () -> FamilyPolicy) {
    private val domainEngine = RuleEngine(policyProvider)
    private val ipEngine = IpRuleEngine(policyProvider)

    fun evaluateDomain(domain: String): FilterDecision {
        val decision = domainEngine.evaluateDomain(domain)
        return FilterDecision(
            action = if (decision.action == RuleEngine.Action.ALLOW) FilterDecision.Action.ALLOW else FilterDecision.Action.BLOCK,
            reason = decision.reason,
            matchedRule = decision.matchedRule
        )
    }

    fun evaluateIpv4(address: String): FilterDecision {
        val decision = ipEngine.evaluateIpv4(address)
        return FilterDecision(
            action = if (decision.action == IpRuleEngine.Action.ALLOW) FilterDecision.Action.ALLOW else FilterDecision.Action.BLOCK,
            reason = decision.reason,
            matchedRule = decision.matchedRule
        )
    }
}
