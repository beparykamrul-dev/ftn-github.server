package com.familytimenet.familyguard.core.filter

/** Normalized result for a metadata-only DNS/IP filtering decision. */
data class FilterDecision(
    val action: Action,
    val reason: String,
    val matchedRule: String? = null
) {
    enum class Action { ALLOW, BLOCK }
}
