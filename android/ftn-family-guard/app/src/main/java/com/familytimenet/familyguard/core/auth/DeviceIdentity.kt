package com.familytimenet.familyguard.core.auth

import android.content.Context
import android.provider.Settings
import java.security.MessageDigest

/** Stable app-scoped pseudonymous device identifier; no hardware serial/IMEI is collected. */
class DeviceIdentity(context: Context) {
    private val appContext = context.applicationContext

    val deviceId: String
        get() {
            val source = Settings.Secure.getString(
                appContext.contentResolver,
                Settings.Secure.ANDROID_ID
            ) ?: "unknown"
            return sha256("ftn-family-guard:$source")
        }

    private fun sha256(value: String): String =
        MessageDigest.getInstance("SHA-256")
            .digest(value.toByteArray(Charsets.UTF_8))
            .joinToString("") { "%02x".format(it) }
}
