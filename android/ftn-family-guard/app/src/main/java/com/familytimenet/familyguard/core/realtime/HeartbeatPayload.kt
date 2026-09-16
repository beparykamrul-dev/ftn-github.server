package com.familytimenet.familyguard.core.realtime

import org.json.JSONObject
import java.util.UUID

/** Small metadata-only heartbeat; no browsing content or packet data is included. */
data class HeartbeatPayload(
    val deviceId: String,
    val policyVersion: Long,
    val appVersion: String,
    val online: Boolean = true,
    val requestId: String = UUID.randomUUID().toString()
) {
    fun toJson(): String = JSONObject()
        .put("type", "device.heartbeat")
        .put("device_id", deviceId)
        .put("policy_version", policyVersion)
        .put("app_version", appVersion)
        .put("online", online)
        .put("request_id", requestId)
        .put("timestamp", System.currentTimeMillis())
        .toString()
}
