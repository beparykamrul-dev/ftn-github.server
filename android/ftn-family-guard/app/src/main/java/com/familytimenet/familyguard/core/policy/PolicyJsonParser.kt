package com.familytimenet.familyguard.core.policy

import com.familytimenet.familyguard.core.model.FamilyPolicy
import org.json.JSONArray
import org.json.JSONObject

/** Parses the FTN policy contract into the local Android policy model. */
object PolicyJsonParser {
    fun parse(json: String): FamilyPolicy {
        val root = JSONObject(json)
        val dns = root.optJSONObject("dns") ?: JSONObject()
        val domains = root.optJSONObject("domains") ?: JSONObject()
        val ips = root.optJSONObject("ips") ?: JSONObject()
        val protection = root.optJSONObject("protection") ?: JSONObject()
        val schedule = root.optJSONObject("schedule") ?: JSONObject()
        val privacy = root.optJSONObject("privacy") ?: JSONObject()
        val enforcement = root.optJSONObject("enforcement") ?: JSONObject()

        return FamilyPolicy(
            version = root.optLong("version", 1L).coerceAtLeast(1L),
            profile = root.optString("profile", "FAMILY").uppercase().let {
                runCatching { FamilyPolicy.Profile.valueOf(it) }.getOrDefault(FamilyPolicy.Profile.FAMILY)
            },
            dns = FamilyPolicy.DnsPolicy(
                enabled = dns.optBoolean("enabled", true),
                encrypted = dns.optBoolean("encrypted", true),
                dnssec = dns.optBoolean("dnssec", true),
                failClosed = dns.optBoolean("fail_closed", false)
            ),
            domains = FamilyPolicy.DomainPolicy(
                allow = strings(domains.optJSONArray("allow")),
                block = strings(domains.optJSONArray("block")),
                categories = strings(domains.optJSONArray("categories"))
            ),
            ips = FamilyPolicy.IpPolicy(
                allowCidrs = strings(ips.optJSONArray("allow_cidrs")),
                blockCidrs = strings(ips.optJSONArray("block_cidrs"))
            ),
            protection = FamilyPolicy.ProtectionPolicy(
                malware = protection.optBoolean("malware", true),
                phishing = protection.optBoolean("phishing", true),
                safeSearch = protection.optBoolean("safe_search", false)
            ),
            schedule = FamilyPolicy.SchedulePolicy(
                enabled = schedule.optBoolean("enabled", false),
                timezone = schedule.optString("timezone", "UTC"),
                allowedFrom = schedule.optStringOrNull("allowed_from"),
                allowedUntil = schedule.optStringOrNull("allowed_until")
            ),
            privacy = FamilyPolicy.PrivacyPolicy(
                rawDnsLogging = privacy.optBoolean("raw_dns_logging", false),
                payloadInspection = privacy.optBoolean("payload_inspection", false),
                aggregateUsageOnly = privacy.optBoolean("aggregate_usage_only", true)
            ),
            enforcement = FamilyPolicy.EnforcementPolicy(
                emergencyBlock = strings(enforcement.optJSONArray("emergency_block")),
                defaultAllow = enforcement.optBoolean("default_allow", true),
                offlineGraceSeconds = enforcement.optLong("offline_grace_seconds", 300L).coerceAtLeast(0L)
            )
        )
    }

    private fun strings(array: JSONArray?): Set<String> = buildSet {
        if (array == null) return@buildSet
        for (i in 0 until array.length()) {
            val value = array.optString(i).trim()
            if (value.isNotEmpty()) add(value)
        }
    }

    private fun JSONObject.optStringOrNull(key: String): String? =
        if (has(key) && !isNull(key)) optString(key).trim().takeIf { it.isNotEmpty() } else null
}
