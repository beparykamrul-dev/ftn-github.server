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
            return when (val type = root.optString("type")) {
                "policy.updated" -> PolicyUpdated(
                    version = root.optLong("version", 0L),
                    policyJson = root.optJSONObject("policy")?.toString()
                        ?: root.optString("policy_json", "")
                )
                "policy.revoked" -> PolicyRevoked(root.optLong("version", 0L))
                "device.command" -> DeviceCommand(
                    command = root.optString("command"),
                    requestId = root.optString("request_id").takeIf { it.isNotBlank() }
                )
                "heartbeat" -> Heartbeat(root.optLong("timestamp", System.currentTimeMillis()))
                "usage.summary" -> UsageSummary(root.optJSONObject("summary")?.toString() ?: "{}")
                else -> Unknown(type)
            }
        }
    }
}
