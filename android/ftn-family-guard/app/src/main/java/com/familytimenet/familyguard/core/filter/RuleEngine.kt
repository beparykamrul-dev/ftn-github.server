package com.familytimenet.familyguard.core.filter

import com.familytimenet.familyguard.core.model.FamilyPolicy

/**
 * DNS/domain decision engine. It deliberately evaluates metadata only;
 * application payloads and message contents are never inspected.
 */
class RuleEngine(private val policyProvider: () -> FamilyPolicy) {
    enum class Action { ALLOW, BLOCK }

    data class Decision(
        val action: Action,
        val reason: String,
        val matchedRule: String? = null
    )

    fun evaluateDomain(input: String): Decision {
        val domain = normalizeDomain(input)
        val policy = policyProvider()

        if (domain.isEmpty()) return Decision(Action.BLOCK, "invalid-domain")

        policy.enforcement.emergencyBlock.firstOrNull { matches(domain, it) }?.let {
            return Decision(Action.BLOCK, "emergency-block", it)
        }
        policy.domains.allow.firstOrNull { matches(domain, it) }?.let {
            return Decision(Action.ALLOW, "explicit-allow", it)
        }
        policy.domains.block.firstOrNull { matches(domain, it) }?.let {
            return Decision(Action.BLOCK, "explicit-block", it)
        }

        return if (policy.enforcement.defaultAllow) {
            Decision(Action.ALLOW, "default-allow")
        } else {
            Decision(Action.BLOCK, "default-block")
        }
    }

    private fun normalizeDomain(value: String): String =
        value.trim().lowercase().removeSuffix(".")

    private fun matches(domain: String, rule: String): Boolean {
        val normalized = normalizeDomain(rule)
        if (normalized.isEmpty()) return false
        if (normalized.startsWith("*.")) {
            val suffix = normalized.removePrefix("*.")
            return domain == suffix || domain.endsWith(".$suffix")
        }
        return domain == normalized || domain.endsWith(".$normalized")
    }
}
