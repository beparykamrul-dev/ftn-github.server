package com.familytimenet.familyguard.core.policy

import java.net.HttpURLConnection
import java.net.URI
import java.net.URLEncoder
import java.util.UUID

/** Acknowledges the exact policy version activated on the Android device. */
class PolicyAckClient(private val baseUrl: String) {
    data class AckResult(val status: Int, val requestId: String)

    fun acknowledge(deviceId: String, sessionReference: String, policyVersion: Long): AckResult {
        require(deviceId.isNotBlank())
        require(sessionReference.isNotBlank())
        require(policyVersion > 0)
        require(URI.create(baseUrl).scheme.equals("https", ignoreCase = true)) {
            "FTN Control Plane must use HTTPS"
        }

        val requestId = UUID.randomUUID().toString()
        val endpoint = URI.create(
            baseUrl.trimEnd('/') + "/api/v1/family/policies/" +
                URLEncoder.encode(deviceId, Charsets.UTF_8.name()).replace("+", "%20") + "/ack"
        )
        val connection = endpoint.toURL().openConnection() as HttpURLConnection
        return try {
            connection.requestMethod = "POST"
            connection.connectTimeout = 10_000
            connection.readTimeout = 10_000
            connection.setRequestProperty("Accept", "application/json")
            connection.setRequestProperty("Content-Type", "application/json")
            connection.setRequestProperty("Authorization", "Bearer $sessionReference")
            connection.setRequestProperty("X-Request-ID", requestId)
            connection.doOutput = true
            val body = "{\"device_id\":\"${escape(deviceId)}\",\"policy_version\":$policyVersion,\"request_id\":\"$requestId\"}"
            connection.outputStream.use { it.write(body.toByteArray(Charsets.UTF_8)) }
            AckResult(connection.responseCode, requestId)
        } finally {
            connection.disconnect()
        }
    }

    private fun escape(value: String): String = buildString {
        value.forEach { ch ->
            when (ch) {
                '\\' -> append("\\\\")
                '"' -> append("\\\"")
                '\n' -> append("\\n")
                '\r' -> append("\\r")
                '\t' -> append("\\t")
                else -> append(ch)
            }
        }
    }
}
