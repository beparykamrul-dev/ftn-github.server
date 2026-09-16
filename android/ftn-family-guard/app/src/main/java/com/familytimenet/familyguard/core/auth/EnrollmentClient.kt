package com.familytimenet.familyguard.core.auth

import android.content.Context
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URI
import java.util.UUID

/** Enrollment transport. Tokens are supplied at runtime and never committed to the repository. */
class EnrollmentClient(context: Context, private val baseUrl: String) {
    private val identity = DeviceIdentity(context.applicationContext)

    data class EnrollmentResult(
        val deviceId: String,
        /** Server-issued opaque session reference; never the enrollment token. */
        val enrollmentId: String,
        val status: Int
    )

    fun enroll(enrollmentToken: String): EnrollmentResult {
        require(enrollmentToken.isNotBlank()) { "Enrollment token is required" }
        val base = URI.create(baseUrl)
        require(base.scheme.equals("https", ignoreCase = true)) {
            "FTN Control Plane must use HTTPS"
        }
        val endpoint = URI.create(base.toString().trimEnd('/') + "/api/v1/family/android/enroll")
        val connection = endpoint.toURL().openConnection() as HttpURLConnection
        return try {
            connection.requestMethod = "POST"
            connection.connectTimeout = 10_000
            connection.readTimeout = 10_000
            connection.setRequestProperty("Accept", "application/json")
            connection.setRequestProperty("Content-Type", "application/json")
            connection.doOutput = true

            val body = """
                {"device_id":"${identity.deviceId}","enrollment_token":"${jsonEscape(enrollmentToken)}","request_id":"${UUID.randomUUID()}"}
            """.trimIndent()
            connection.outputStream.use { it.write(body.toByteArray(Charsets.UTF_8)) }

            val status = connection.responseCode
            val stream = if (status in 200..299) connection.inputStream else connection.errorStream
            val responseText = stream?.bufferedReader(Charsets.UTF_8)?.use { it.readText() }.orEmpty()
            if (status !in 200..299) {
                throw IllegalStateException("Family Guard enrollment failed: HTTP $status")
            }
            val response = runCatching { JSONObject(responseText) }.getOrElse {
                throw IllegalStateException("Family Guard returned invalid enrollment response")
            }
            val deviceId = response.optString("device_id").ifBlank { identity.deviceId }
            val session = response.optString("session").trim()
            require(session.isNotBlank()) { "Family Guard did not return a session reference" }
            EnrollmentResult(deviceId = deviceId, enrollmentId = session, status = status)
        } finally {
            connection.disconnect()
        }
    }

    private fun jsonEscape(value: String): String = buildString {
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
