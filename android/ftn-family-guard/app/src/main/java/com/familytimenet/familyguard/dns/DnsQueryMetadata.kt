package com.familytimenet.familyguard.dns

/** Minimal DNS question metadata; no application payload or answer content is retained. */
data class DnsQueryMetadata(
    val transactionId: Int,
    val domain: String,
    val recordType: Int,
    val recordClass: Int
)
