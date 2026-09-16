package com.familytimenet.familyguard.core.policy

import java.net.HttpURLConnection
import java.net.URI

/** Fetches and parses the versioned Family Guard policy over HTTPS. */
class PolicyClient(private val baseUrl: String) {
    fun fetchPolicy(deviceId: String, sessionReference: String): PolicyEnvelope {
        require(deviceId.isNotBlank())
        require(sessionReference.isNotBlank())
        require(URI.create(baseUrl).scheme.equals("https", ignoreCase = true)) {
            "FTN Control Plane must use HTTPS"
        }

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

            val status = connection.responseCode
            if (status !in 200..299) return PolicyEnvelope(status, null, null)

            val rawJson = connection.inputStream.bufferedReader(Charsets.UTF_8).use { it.readText() }
            PolicyEnvelope(status, PolicyJsonParser.parse(rawJson), rawJson)
        } finally {
            connection.disconnect()
        }
    }

    data class PolicyEnvelope(
        val status: Int,
        val policy: com.familytimenet.familyguard.core.model.FamilyPolicy?,
        val rawJson: String?
    )

    private fun encodePath(value: String): String =
        java.net.URLEncoder.encode(value, Charsets.UTF_8.name()).replace("+", "%20")
}
