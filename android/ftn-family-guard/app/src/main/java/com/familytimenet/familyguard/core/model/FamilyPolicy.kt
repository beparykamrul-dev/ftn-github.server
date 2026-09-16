package com.familytimenet.familyguard.core.model

/** Local, versioned policy used by FTN Family Guard. */
data class FamilyPolicy(
    val version: Long,
    val profile: Profile,
    val dns: DnsPolicy,
    val domains: DomainPolicy,
    val ips: IpPolicy,
    val protection: ProtectionPolicy,
    val schedule: SchedulePolicy,
    val privacy: PrivacyPolicy,
    val enforcement: EnforcementPolicy
) {
    enum class Profile { FAMILY, CHILD, TEEN, CUSTOM }

    data class DnsPolicy(
        val enabled: Boolean = true,
        val encrypted: Boolean = true,
        val dnssec: Boolean = true,
        val failClosed: Boolean = false
    )

    data class DomainPolicy(
        val allow: Set<String> = emptySet(),
        val block: Set<String> = emptySet(),
        val categories: Set<String> = emptySet()
    )

    data class IpPolicy(
        val allowCidrs: Set<String> = emptySet(),
        val blockCidrs: Set<String> = emptySet()
    )

    data class ProtectionPolicy(
        val malware: Boolean = true,
        val phishing: Boolean = true,
        val safeSearch: Boolean = false
    )

    data class SchedulePolicy(
        val enabled: Boolean = false,
        val timezone: String = "UTC",
        val allowedFrom: String? = null,
        val allowedUntil: String? = null
    )

    data class PrivacyPolicy(
        val rawDnsLogging: Boolean = false,
        val payloadInspection: Boolean = false,
        val aggregateUsageOnly: Boolean = true
    )

    data class EnforcementPolicy(
        val emergencyBlock: Set<String> = emptySet(),
        val defaultAllow: Boolean = true,
        val offlineGraceSeconds: Long = 300
    )

    companion object {
        fun default(): FamilyPolicy = FamilyPolicy(
            version = 1L,
            profile = Profile.FAMILY,
            dns = DnsPolicy(),
            domains = DomainPolicy(),
            ips = IpPolicy(),
            protection = ProtectionPolicy(),
            schedule = SchedulePolicy(),
            privacy = PrivacyPolicy(),
            enforcement = EnforcementPolicy()
        )
    }
}
