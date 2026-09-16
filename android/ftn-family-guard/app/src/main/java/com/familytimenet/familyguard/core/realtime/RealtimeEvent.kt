package com.familytimenet.familyguard.core.realtime

import org.json.JSONObject

/** Normalized realtime events consumed by policy, device and usage layers. */
sealed interface RealtimeEvent {
    data class PolicyUpdated(val version: Long, val policyJson: String) : RealtimeEvent
    data class PolicyRevoked(val version: Long) : RealtimeEvent
    data class DeviceCommand(val command: String, val requestId: String?) : RealtimeEvent
    data class Heartbeat(val timestamp: Long) : RealtimeEvent
    data class UsageSummary(val json: String) : RealtimeEvent
    data class Unknown(val type: String) : RealtimeEvent

    companion object {
        fun parse(text: String): RealtimeEvent {
            val root = JSONObject(text)
            val type = root.optString("type").ifBlank { root.optString("event") }
            val payload = root.optJSONObject("payload") ?: root
            return when (type) {
                "policy.updated" -> PolicyUpdated(payload.optLong("version", root.optLong("version", 0L)), payload.optJSONObject("policy")?.toString() ?: payload.optString("policy_json", ""))
                "policy.revoked" -> PolicyRevoked(payload.optLong("version", root.optLong("version", 0L)))
                "device.command" -> DeviceCommand(payload.optString("command"), payload.optString("request_id").takeIf { it.isNotBlank() })
                "heartbeat" -> Heartbeat(payload.optLong("timestamp", root.optLong("timestamp", System.currentTimeMillis())))
                "usage.summary" -> UsageSummary(payload.optJSONObject("summary")?.toString() ?: payload.toString())
                else -> Unknown(type)
            }
        }
    }
}
