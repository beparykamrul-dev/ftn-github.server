package com.familytimenet.familyguard.core.filter

import com.familytimenet.familyguard.core.model.FamilyPolicy

/** Metadata-only IPv4/CIDR decision engine. It never inspects application payloads. */
class IpRuleEngine(private val policyProvider: () -> FamilyPolicy) {
    enum class Action { ALLOW, BLOCK }

    data class Decision(val action: Action, val reason: String, val matchedRule: String? = null)

    fun evaluateIpv4(address: String): Decision {
        val ip = parseIpv4(address) ?: return Decision(Action.BLOCK, "invalid-ip")
        val policy = policyProvider()

        policy.ips.allowCidrs.firstOrNull { contains(ip, it) }?.let {
            return Decision(Action.ALLOW, "explicit-allow-cidr", it)
        }
        policy.ips.blockCidrs.firstOrNull { contains(ip, it) }?.let {
            return Decision(Action.BLOCK, "explicit-block-cidr", it)
        }
        return if (policy.enforcement.defaultAllow) {
            Decision(Action.ALLOW, "default-allow")
        } else {
            Decision(Action.BLOCK, "default-block")
        }
    }

    private fun contains(ip: Long, cidr: String): Boolean {
        val parts = cidr.trim().split('/', limit = 2)
        val network = parseIpv4(parts[0]) ?: return false
        val prefix = parts.getOrNull(1)?.toIntOrNull() ?: 32
        if (prefix !in 0..32) return false
        val mask = if (prefix == 0) 0L else (0xFFFFFFFFL shl (32 - prefix)) and 0xFFFFFFFFL
        return (ip and mask) == (network and mask)
    }

    private fun parseIpv4(value: String): Long? {
        val parts = value.trim().split('.')
        if (parts.size != 4) return null
        var result = 0L
        for (part in parts) {
            val octet = part.toIntOrNull() ?: return null
            if (octet !in 0..255) return null
            result = (result shl 8) or octet.toLong()
        }
        return result
    }
}
