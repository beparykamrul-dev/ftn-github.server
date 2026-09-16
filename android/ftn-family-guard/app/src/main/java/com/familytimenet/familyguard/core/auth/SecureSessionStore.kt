package com.familytimenet.familyguard.core.auth

import android.content.Context
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.sessionDataStore by preferencesDataStore(name = "ftn_secure_session")

/** Stores opaque session references locally; secrets are never written to Git. */
class SecureSessionStore(private val context: Context) {
    private val sessionKey = stringPreferencesKey("session_reference")
    private val enrollmentKey = stringPreferencesKey("enrollment_id")

    val sessionReference: Flow<String?> =
        context.sessionDataStore.data.map { it[sessionKey] }

    val enrollmentId: Flow<String?> =
        context.sessionDataStore.data.map { it[enrollmentKey] }

    suspend fun save(sessionReference: String, enrollmentId: String) {
        require(sessionReference.isNotBlank())
        require(enrollmentId.isNotBlank())
        context.sessionDataStore.updateData { current ->
            current.toMutablePreferences().apply {
                this[sessionKey] = sessionReference
                this[enrollmentKey] = enrollmentId
            }
        }
    }

    suspend fun clear() {
        context.sessionDataStore.updateData { current ->
            current.toMutablePreferences().apply {
                remove(sessionKey)
                remove(enrollmentKey)
            }
        }
    }
}
