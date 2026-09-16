package com.familytimenet.familyguard.core.usage

import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.atomic.AtomicLong

/** Aggregates filter outcomes without retaining raw DNS queries or payloads. */
class UsageAggregator {
    private val allowed = AtomicLong()
    private val blocked = AtomicLong()
    private val failed = AtomicLong()
    private val categories = ConcurrentHashMap<String, AtomicLong>()

    fun recordAllowed(category: String? = null) {
        allowed.incrementAndGet()
        recordCategory(category)
    }

    fun recordBlocked(category: String? = null) {
        blocked.incrementAndGet()
        recordCategory(category)
    }

    fun recordFailed() {
        failed.incrementAndGet()
    }

    fun snapshot(deviceId: String, windowStart: Long, windowEnd: Long): UsageSnapshot =
        UsageSnapshot(
            deviceId = deviceId,
            windowStart = windowStart,
            windowEnd = windowEnd,
            allowed = allowed.get(),
            blocked = blocked.get(),
            failed = failed.get(),
            topCategories = categories.entries
                .map { it.key to it.value.get() }
                .sortedByDescending { it.second }
                .take(10)
                .toMap()
        )

    fun reset() {
        allowed.set(0)
        blocked.set(0)
        failed.set(0)
        categories.clear()
    }

    private fun recordCategory(category: String?) {
        val normalized = category?.trim()?.lowercase()?.takeIf { it.isNotEmpty() } ?: return
        categories.computeIfAbsent(normalized) { AtomicLong() }.incrementAndGet()
    }

    data class UsageSnapshot(
        val deviceId: String,
        val windowStart: Long,
        val windowEnd: Long,
        val allowed: Long,
        val blocked: Long,
        val failed: Long,
        val topCategories: Map<String, Long>
    )
}
