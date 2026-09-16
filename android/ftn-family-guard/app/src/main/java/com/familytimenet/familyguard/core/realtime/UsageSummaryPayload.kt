package com.familytimenet.familyguard.core.realtime

import org.json.JSONArray
import org.json.JSONObject

/** Aggregated DNS usage metadata. Raw queries, payloads and message contents are excluded. */
data class UsageSummaryPayload(
    val deviceId: String,
    val windowStart: Long,
    val windowEnd: Long,
    val allowed: Long,
    val blocked: Long,
    val failed: Long,
    val topCategories: Map<String, Long> = emptyMap()
) {
    fun toJson(): String {
        val categories = JSONArray()
        topCategories.entries.forEach { (category, count) ->
            categories.put(JSONObject().put("category", category).put("count", count))
        }
        return JSONObject()
            .put("type", "usage.summary")
            .put("device_id", deviceId)
            .put("window_start", windowStart)
            .put("window_end", windowEnd)
            .put("allowed", allowed)
            .put("blocked", blocked)
            .put("failed", failed)
            .put("top_categories", categories)
            .toString()
    }
}
