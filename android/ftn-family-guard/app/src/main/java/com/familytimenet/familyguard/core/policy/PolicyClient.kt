package com.familytimenet.familyguard.core.policy

import com.familytimenet.familyguard.core.model.FamilyPolicy
import java.net.HttpURLConnection
import java.net.URI

/** Read-only policy transport. Server response parsing is intentionally kept behind a small contract. */
class PolicyClient(private val baseUrl: String) {
    fun fetchPolicy(deviceId: String, sessionReference: String): PolicyEnvelope {
        require(deviceId.isNotBlank())
        require(sessionReference.isNotBlank())

        val endpoint = URI.create(
            baseUrl.trimEnd('/') + "/api/v1/family/policies/" + encodePath(deviceId)
        )
        val connection = endpoint.toURL().openConnection() as HttpURLConnection
        return try {
            connection.requestMethod = "GET"
            connection.connectTimeout = 10_000
            connection.readTimeout = 10_000
            connection.setRequestProperty("Accept", "application/json")
            connection.setRequestProperty("Authorization", "Bearer $sessionReference")

            PolicyEnvelope(
                status = connection.responseCode,
                policy = if (connection.responseCode in 200..299) FamilyPolicy.default() else null
            )
        } finally {
            connection.disconnect()
        }
    }

    data class PolicyEnvelope(val status: Int, val policy: FamilyPolicy?)

    private fun encodePath(value: String): String =
        java.net.URLEncoder.encode(value, Charsets.UTF_8.name()).replace("+", "%20")
}
