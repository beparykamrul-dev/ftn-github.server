package com.familytimenet.familyguard.core.auth

import android.content.Context
import java.net.HttpURLConnection
import java.net.URI
import java.util.UUID

/** Enrollment transport. Tokens are supplied at runtime and never committed to the repository. */
class EnrollmentClient(context: Context, private val baseUrl: String) {
    private val identity = DeviceIdentity(context.applicationContext)

    data class EnrollmentResult(
        val deviceId: String,
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

            EnrollmentResult(
                deviceId = identity.deviceId,
                enrollmentId = UUID.randomUUID().toString(),
                status = connection.responseCode
            )
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
